package game

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/mreliasen/swi-server/internal/responses"
)

type Location struct {
	City        *City
	Description string
	Coords      Coordinates
	Players     map[*Client]bool
	Npcs        map[*Entity]bool
	Items       map[*Item]bool
	PlayerJoin  chan *Client
	NpcJoin     chan *Entity
	AddItem     chan *ItemMoved
	RemoveItem  chan *ItemMoved
	Events      chan *ClientResponse
	Respawn     chan CombatAction
	Buildings   []*Building
	mu          sync.Mutex
}

type Movement struct {
	Origin      Coordinates `json:"origin,omitempty"`
	Destination Coordinates `json:"destination,omitempty"`
}

func (l *Location) buildingGameFrames() []*responses.Building {
	frames := []*responses.Building{}

	for _, b := range l.Buildings {
		frames = append(frames, b.gameFrame())
	}

	return frames
}

func (l *Location) itemsGameFrames() []*responses.Item {
	frames := []*responses.Item{}

	for itm := range l.Items {
		frames = append(frames, itm.GameFrame())
	}

	return frames
}

func (l *Location) getBuildingCommand(cmdKey string) (cmd *Command, exists bool) {
	for _, b := range l.Buildings {
		if val, ok := b.Commands[cmdKey]; ok {
			return val, true
		}
	}

	return nil, false
}

func (l *Location) run() {
	go func() {
		for {
			action := <-l.Respawn
			l.mu.Lock()
			l.Players[action.Target.Client] = true
			l.mu.Unlock()

			player := action.Target.Client.Player
			player.Mu.Lock()

			if player.Loc != nil {
				origin := player.Loc
				origin.mu.Lock()
				delete(origin.Players, action.Target.Client)
				origin.mu.Unlock()
			}

			player.Loc = l
			player.Dead = false
			player.Mu.Unlock()

			// Notify the new location
			go player.sendGameFrame(false)
			go player.PlayerSendMapUpdate()
			go player.PlayerSendStatsUpdate()
			go player.PlayerSendInventoryUpdate()
			go player.PlayerSendLocationUpdate()

			action.Target.Client.SendEvent(&responses.Generic{
				Status: responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{
					fmt.Sprintf("%s put you down on the ground. Luckily some bystanders called an amulance which picked up before bleeding out.", action.Attacker.Name),
					"You walk out of the hospital after the doctors patched up up, mostly.",
				},
			})

			event := CreateEvent(map[uint64]bool{action.Target.Client.Player.PlayerID: true}, &responses.Generic{
				Messages: []string{
					fmt.Sprintf("You see %s walk out of the hospital.", action.Target.Name),
				},
			})

			l.Events <- &event
		}
	}()

	go func() {
		for {
			client := <-l.PlayerJoin
			l.mu.Lock()
			l.Players[client] = true
			l.mu.Unlock()

			fled := false
			if len(client.Player.TargetedBy) > 0 {
				client.Player.RemoveTargetLock()
				fled = true
			}

			var origin *Location
			movement := Movement{
				Destination: l.Coords,
			}

			hasOrigin := client.Player.Loc != nil
			sameCity := true

			if hasOrigin {
				origin = client.Player.Loc
				movement.Origin = origin.Coords
				// remove player from origin
				origin.mu.Lock()
				delete(origin.Players, client)
				origin.mu.Unlock()
				sameCity = origin.City.ShortName == l.City.ShortName

				if !sameCity {
					origin.City.PlayerLeave <- client
				}
			}

			client.Player.Mu.Lock()
			client.Player.Loc = l
			client.Player.LastLocation = Coordinates{
				North: l.Coords.North,
				East:  l.Coords.East,
				City:  l.City.ShortName,
			}
			client.Player.Mu.Unlock()

			event := CreateEvent(map[uint64]bool{client.Player.PlayerID: true}, &responses.PlayerMoveEvent{
				Type:      responses.MoveEventType_MOVE_EVENT_ARRIVE,
				Player:    client.Player.PlayerGameFrame(),
				Direction: getFromDirection(origin, l),
				Samecity:  sameCity,
				Fled:      fled,
			})

			// Notify the new location
			go client.Player.sendGameFrame(false)
			go client.Player.PlayerSendMapUpdate()
			l.Events <- &event

			if fled {
				client.SendEvent(&responses.Generic{
					Messages: []string{
						"You managed to escape from the battle. however the news of such a cowardly act spreads fast. In the scuttle to escape, you dropped some items.",
					},
				})

				for user := range origin.Players {
					user.Player.sendGameFrame(false)
				}
			}

			if !sameCity {
				client.SendEvent(&responses.Generic{
					Messages: []string{
						fmt.Sprintf("You land in %s", l.City.Name),
					},
				})
			}

			if hasOrigin {
				event = CreateEvent(map[uint64]bool{client.Player.PlayerID: true}, &responses.PlayerMoveEvent{
					Type:      responses.MoveEventType_MOVE_EVENT_LEAVE,
					Player:    client.Player.PlayerGameFrame(),
					Direction: getToDirection(origin, l),
					Samecity:  sameCity,
					Fled:      fled,
				})

				// notifty the old location
				origin.Events <- &event
			}

			if _, ok := l.City.Players[client]; !ok {
				l.City.PlayerJoin <- client
			}
		}
	}()

	go func() {
		for {
			npc := <-l.NpcJoin
			l.mu.Lock()
			l.Npcs[npc] = true
			l.mu.Unlock()

			var origin *Location
			hasOrigin := npc.Loc != nil
			movement := Movement{
				Destination: l.Coords,
			}

			if hasOrigin {
				origin = npc.Loc
				movement.Origin = origin.Coords
				// remove player from origin
				origin.mu.Lock()
				delete(origin.Npcs, npc)
				origin.mu.Unlock()
			}

			npc.Mu.Lock()
			npc.Loc = l
			npc.Mu.Unlock()

			val := CreateEvent(nil, &responses.NPCMoveEvent{
				Type:      responses.MoveEventType_MOVE_EVENT_ARRIVE,
				Npc:       npc.NPCGameFrame(),
				Direction: getFromDirection(origin, l),
			})

			// notifty the new location
			if len(l.Players) > 0 {
				l.Events <- &val
			}

			if !hasOrigin {
				continue
			}

			// notifty the old location
			if len(origin.Players) > 0 {
				val = CreateEvent(nil, &responses.NPCMoveEvent{
					Type:      responses.MoveEventType_MOVE_EVENT_LEAVE,
					Npc:       npc.NPCGameFrame(),
					Direction: getToDirection(origin, l),
				})

				origin.Events <- &val
			}
		}
	}()

	go func() {
		for {
			// item dropped
			event := <-l.AddItem
			l.mu.Lock()
			event.Item.Loc = l
			l.Items[event.Item] = true
			l.mu.Unlock()

			if event.Player != nil {
				event.Player.Client.SendEvent(&responses.Generic{
					Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
					Messages: []string{fmt.Sprintf("You toss %s on the ground.", event.Item.InspectName())},
				})

				// only show this if it was not an NPC dying
				if len(l.Players) > 0 {
					val := CreateEvent(map[uint64]bool{event.Player.PlayerID: true}, &responses.Generic{
						Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
						Messages: []string{fmt.Sprintf("You see %s toss %s on the ground.", event.By, event.Item.InspectName())},
					})

					l.Events <- &val
				}
			}

			for p := range l.Players {
				p.Player.sendGameFrame(false)
			}
		}
	}()

	go func() {
		for {
			// item picked up
			event := <-l.RemoveItem
			if event.Player == nil {
				continue
			}

			l.mu.Lock()
			if ok := l.Items[event.Item]; !ok {
				event.Player.Client.SendEvent(&responses.Generic{
					Status:   responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{"That item is no longere there."},
				})

				l.mu.Unlock()
				continue
			}

			err := event.Player.Inventory.addItem(event.Item)
			if err != nil {
				event.Player.Client.SendEvent(&responses.Generic{
					Status:   responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{"You do not have room in your inventory"},
				})

				l.mu.Unlock()
				continue
			}

			event.Player.PlayerSendInventoryUpdate()
			delete(l.Items, event.Item)
			l.mu.Unlock()

			if len(l.Players) > 0 {
				event.Player.sendGameFrame(false)

				val := CreateEvent(map[uint64]bool{event.Player.PlayerID: true}, &responses.Generic{
					Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
					Messages: []string{fmt.Sprintf("%s picked up %s from the ground.", event.By, event.Item.InspectName())},
				})

				l.Events <- &val
			}

		}
	}()

	go func() {
		for {
			// new local chat
			event := <-l.Events

			for client := range l.Players {
				if event.Ignore != nil {
					if ok := event.Ignore[client.Player.PlayerID]; ok {
						continue
					}
				}

				go client.Player.sendGameFrame(false)

				select {
				case client.Send <- event.Payload:
				default:
					delete(l.Players, client)
				}
			}
		}
	}()
}

func getToDirection(from *Location, to *Location) responses.Direction {
	switch {
	case from == nil:
		return responses.Direction_DIRECTION_UNKNOWN
	case to == nil:
		return responses.Direction_DIRECTION_UNKNOWN
	case to.Coords.North > from.Coords.North:
		return responses.Direction_DIRECTION_NORTH
	case to.Coords.North < from.Coords.North:
		return responses.Direction_DIRECTION_SOUTH
	case to.Coords.East > from.Coords.East:
		return responses.Direction_DIRECTION_EAST
	case to.Coords.East < from.Coords.East:
		return responses.Direction_DIRECTION_WEST
	}

	return responses.Direction_DIRECTION_UNKNOWN
}

func getFromDirection(from *Location, to *Location) responses.Direction {
	switch {
	case from == nil:
		return responses.Direction_DIRECTION_UNKNOWN
	case to == nil:
		return responses.Direction_DIRECTION_UNKNOWN
	case to.Coords.North > from.Coords.North:
		return responses.Direction_DIRECTION_SOUTH
	case to.Coords.North < from.Coords.North:
		return responses.Direction_DIRECTION_NORTH
	case to.Coords.East > from.Coords.East:
		return responses.Direction_DIRECTION_WEST
	case to.Coords.East < from.Coords.East:
		return responses.Direction_DIRECTION_EAST
	}

	return responses.Direction_DIRECTION_UNKNOWN
}

func CreateLocation(c *City, north int, east int) *Location {
	loc := Location{
		City: c,
		Coords: Coordinates{
			North: north,
			East:  east,
		},
		Description: locationDescriptions[rand.Intn(len(locationDescriptions))],
		Players:     make(map[*Client]bool),
		Npcs:        make(map[*Entity]bool),
		Items:       make(map[*Item]bool),
		PlayerJoin:  make(chan *Client),
		NpcJoin:     make(chan *Entity),
		AddItem:     make(chan *ItemMoved),
		RemoveItem:  make(chan *ItemMoved),
		Events:      make(chan *ClientResponse),
		Respawn:     make(chan CombatAction),
		Buildings:   []*Building{},
	}

	for _, poi := range c.BuildingLocations {
		if poi.Coords.North == loc.Coords.North && poi.Coords.East == loc.Coords.East {
			for _, bType := range poi.Buildings {
				building := NewBuilding(bType)
				loc.Buildings = append(loc.Buildings, &building)
			}
		}
	}

	go loc.run()
	return &loc
}

var locationDescriptions = []string{
	"Broken lamp posts line the street.",
	"You hear a city bus in the distance.",
	"Broken glass covers the road.",
	"Music from a distant strip bar echoes through the street.",
	"Shouts from a nearby complication can be easily heard.",
	"Garbage lines the city gutters",
	"You're standing outside an abandoned building.",
	"You're standing next to a neglected building.",
	"You hear rap music thumping in the distance.",
	"Beer bottles clutter the sidewalk.",
	"You hear police sirens in the distance.",
	"You see trash blowing down the street.",
	"You stand outside a closed boarded up shop.",
	"Vandalised advertisements cover the sidewalk walls.",
	"A fire burns in a drum nearby.",
	"You see an abandoned car on with no tires and broken windows.",
	"You see rubble everywhere.",
	"Graffiti covers the city walls.",
	"You're standing in a Dark Alley.",
	"The flickering neon lights create an eerie atmosphere.",
	"You spot a payphone, its receiver hanging off the hook.",
	"A group of pigeons takes flight as you approach.",
	"The distant hum of a subway train rumbles through the city.",
	"A stray dog scavenges through a pile of discarded fast-food wrappers.",
	"The sound of breakdancing music carries from a street performer nearby.",
	"An overgrown, neglected park stretches before you.",
	"You notice a torn-up newspaper, its headlines barely legible.",
	"A group of kids skateboarding in an empty pool catches your eye.",
	"A flickering streetlight buzzes and casts eerie shadows on the sidewalk.",
	"An abandoned shopping cart sits by a graffiti-covered wall.",
	"Rain-soaked cardboard boxes are scattered along the pavement.",
	"The distant aroma of roasted chestnuts fills the air.",
	"You find a '90s-era boombox discarded on the sidewalk.",
	"A gust of wind kicks up a swirl of fallen leaves.",
	"You come across a tagged-up paywall with 'No Justice, No Peace' spray-painted on it.",
	"You spot a poster advertising a forgotten rock concert.",
	"The buzzing neon sign of a 24-hour diner catches your attention.",
	"An abandoned newsstand remains untouched for years.",
	"A flickering 'Open' sign dangles precariously over a worn-out diner.",
	"You come across a basketball court, its surface cracked and worn.",
	"A newspaper blows past, its headline screaming about a Y2K scare.",
	"You see an old poster for a '90s action movie on a telephone pole.",
	"An alley cat darts out of sight as you approach.",
	"You stumble upon an abandoned arcade with fading 'Game Over' signs.",
	"A vintage clothing store's mannequins stand frozen in time.",
	"You see an old CRT TV set discarded on the sidewalk.",
	"A poster advertises a rave from the late '90s, almost torn apart.",
	"A handwritten sign points to 'Vinyl Records' in an alleyway.",
	"The distant hum of a late-night talk show host drifts from an open window.",
	"You notice an old phone booth, its glass cracked and graffitied.",
}
