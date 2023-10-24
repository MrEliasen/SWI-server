package game

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/mreliasen/swi-server/internal/responses"
	"github.com/pterm/pterm"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Game struct {
	DbConn       *sql.DB                        // DB Connection
	Logout       chan *Client                   // disconnected clients, gracefull logout
	Clients      map[*Client]bool               // connected clients
	Players      map[*Entity]bool               // connected clients
	Register     chan *Client                   // pending new connections
	GlobalEvents chan protoreflect.ProtoMessage // global chat
	NewsFlash    chan protoreflect.ProtoMessage // news flashes
	Movement     chan protoreflect.ProtoMessage // player movement events
	World        map[string]*City               // the game world
	mu           sync.Mutex
}

func (g *Game) SendMOTD(c *Client) {
	c.SendEvent(&responses.Generic{
		Ascii: true,
		Messages: []string{
			"   _____ _                 _    __          __               _____            ",
			"  / ____| |               | |   \\ \\        / /              |_   _|           ",
			" | (___ | |_ _ __ ___  ___| |_   \\ \\  /\\  / /_ _ _ __ ___     | |  _ __   ___ ",
			"  \\___ \\| __| '__/ _ \\/ _ \\ __|   \\ \\/  \\/ / _` | '__/ __|    | | | '_ \\ / __|",
			"  ____) | |_| | |  __/  __/ |_     \\  /\\  / (_| | |  \\__ \\   _| |_| | | | (__ ",
			" |_____/ \\__|_|  \\___|\\___|\\__|     \\/  \\/ \\__,_|_|  |___/  |_____|_| |_|\\___|",
			" ==================================================================================== ",
			"Welcome to Street Wars Inc. (Beta)",
			"A tribute to the original Street Wars Online 2 from 2000",
			"If you have an account type /login, otherwise /register",
			"To Learn about the game controls to /help",
			"This is an unstable-beta version of the game, expect crashes and other issues.",
			"If you experience any bugs or issues, please report them to mark.eliasen@pm.me",
			"There is no official Discord or Reddit or similar.",
		},
	})
}

func (g *Game) GetPlayerClient(playerId uint64) *Entity {
	for p := range g.Players {
		if p.PlayerID == playerId {
			return p
		}
	}

	return nil
}

func (g *Game) LoginPlayer(c *Client, p *Entity, k *Coordinates) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var existingPlayer *Entity
	for currentPlayer := range g.Players {
		if currentPlayer.PlayerID == p.PlayerID {
			existingPlayer = p
			break
		}
	}

	if existingPlayer != nil {
		c.SendEvent(&responses.Generic{
			Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
			Messages: []string{
				fmt.Sprintf("This player is already logged into the game. If you disconnected while being aim locked your character stays in-game for %d seconds.", settings.CombatLoggingSecs),
			},
		})

		c.Mu.Lock()
		c.Authenticated = false
		c.UserId = 0
		c.Mu.Unlock()
		return
	}

	c.Player = p
	p.Client = c
	g.Players[p] = true
	c.Player.Inventory.load()
	p.StartRoutines()

	c.Game.Register <- c

	c.SendEvent(&responses.System{
		Type: responses.SystemType_GAME_READY,
	})

	go p.PlayerSendInventoryUpdate()
	go p.PlayerSendStatsUpdate()
	go p.PlayerSendPlayerList()

	var loc *Location

	if k != nil {
		if _, ok := p.Client.Game.World[k.City]; ok {
			loc = g.World[k.City].Grid[k.toString()]
		}
	}

	if loc == nil {
		city := g.World[p.Hometown]
		coords := city.RandomLocation()
		loc = city.Grid[coords.toString()]
	}

	event := &responses.PlayerList{
		Type:     responses.PlayerEvent_EVENT_TYPE_PLAYER_JOIN,
		Id:       c.UUID,
		Name:     c.Player.Name,
		Hometown: c.Player.Hometown,
		GangTag:  c.Player.GangTag(),
	}

	c.Send <- event

	loc.PlayerJoin <- c
	g.GlobalEvents <- event

	c.SendEvent(&responses.Generic{
		Ascii: true,
		Messages: []string{
			"If you are new to the game, here is some pointers:",
			"- Your skills increase by using them (like Accuracy, inceases in combat).",
			"- To attack an NPC or player, you must first take /aim on them.",
			"- See /help for controls",
			"- See /npcs for list of NPCs and more",
			"- Dealing drugs is the primary way to gain money and early rep.",
			"- Buy from drug dealers with /buy and sell to drug addicts with /sell",
			"- Careful which NPCs you attack, if you get into trouble you can \"/flee <direction>\"",
			"- You can buy a smart phone at the Pawn Shop once you have enough money, enables GPS.",
		},
	})
}

func (g *Game) Restock() {
	for _, city := range g.World {
		for npc := range city.NPCs[DrugDealer] {
			npc.Restock()
		}

		for npc := range city.NPCs[DrugAddict] {
			npc.ClearStock()
		}

		city.UpdateDrugDemand()
	}

	g.GlobalEvents <- &responses.NewsFlash{
		Msg: "<NEWS FLASH> Word on the street says that new shipments of illegal drugs has hit all major cities.",
	}
}

func (g *Game) Save() {
	for client := range g.Clients {
		if client.Player != nil {
			client.Player.Save()
		}
	}
}

func (g *Game) Run() {
	go func() {
		for {
			for _, city := range g.World {
				city.RandomiseTravelCost()
			}

			time.Sleep(settings.TravelCostChangeMinutes * time.Minute)
		}
	}()

	go func() {
		for {
			time.Sleep(settings.AutoSaveMinutes * time.Minute)
			g.Save()
		}
	}()

	go func() {
		for {
			// drug refreshing
			g.Restock()
			time.Sleep(settings.DrugRestockDelaySeconds * time.Second)
		}
	}()

	go func() {
		for {
			// new player
			client := <-g.Register
			g.mu.Lock()
			g.Clients[client] = true
			g.mu.Unlock()
		}
	}()

	go func() {
		for {
			message := <-g.GlobalEvents
			for client := range g.Clients {
				select {
				case client.Send <- message:
				default:
					g.mu.Lock()
					delete(g.Clients, client)
					g.mu.Unlock()
				}
			}
		}
	}()

	go func() {
		for {
			news := <-g.NewsFlash
			for client := range g.Clients {
				select {
				case client.Send <- news:
				default:
					g.mu.Lock()
					delete(g.Clients, client)
					g.mu.Unlock()
				}
			}
		}
	}()

	go func() {
		for {
			// graceful logout player
			client := <-g.Logout

			client.Mu.Lock()
			g.mu.Lock()
			delete(g.Clients, client)
			g.mu.Unlock()

			if client.Player != nil {
				if len(client.Player.TargetedBy) > 0 {
					client.Player.Save()
					logger.LogCombat("", client.Player.Name, "combatlogging", "", 0, 0, 0, "")
					logger.Logger.Info(fmt.Sprintf("%s combat logging.", client.Player.Name))
					client.CombatLogging = true
				}
			}

			client.Mu.Unlock()

			go func() {
				if client.CombatLogging {
					time.Sleep(settings.CombatLoggingSecs * time.Second)
				}

				g.HandleLogout(client)
			}()
		}
	}()
}

func (g *Game) HandleLogout(client *Client) {
	name := "<Guest>"
	logger.Logger.Trace("Handling logout")

	if client.Player != nil {
		player := client.Player
		player.Save()
		logger.Logger.Trace("saving logged out player")
		name = player.Name

		if len(player.TargetedBy) > 0 {
			player.RemoveTargetLock()
		}

		if player.Loc != nil {
			logger.Logger.Trace("locking player")
			player.Loc.mu.Lock()
			delete(player.Loc.City.Players, client)
			delete(player.Loc.Players, client)
			player.Loc.mu.Unlock()
			logger.Logger.Trace("unlocking player")
		}

		if player.ShoppingWith != nil {
			if len(player.ShoppingWith) > 0 {
				logger.Logger.Trace("locking shoppers")
				for shopper := range player.ShoppingWith {
					shopper.Mu.Lock()
					delete(shopper.ShoppingWith, player)
					shopper.Mu.Unlock()
				}
				logger.Logger.Trace("unlocking shoppers")
			}
		}

		g.GlobalEvents <- &responses.PlayerList{
			Type: responses.PlayerEvent_EVENT_TYPE_PLAYER_LEAVE,
			Id:   client.UUID,
		}

		g.mu.Lock()
		delete(g.Players, player)
		g.mu.Unlock()
	}

	logger.Logger.Info(pterm.Sprintf("%s, Logged out", name))
}

func (g *Game) RenderConsoleUI() {
	area, _ := pterm.DefaultArea.Start()

	for {
		list := [][]string{
			{"Player", "Location", "Coords"},
		}

		playersOnline := len(g.Clients)

		if playersOnline > 0 {
			for client := range g.Clients {
				name := "<Waiting>"
				city := "-"
				coords := "-"

				if client.Player != nil {
					loc := client.Player.Loc

					if loc != nil {
						city = loc.City.ShortName
						coords = loc.Coords.toString()
						name = client.Player.Name
					}
				}

				if client.CombatLogging {
					name = "(CL) " + name
				}

				list = append(list, []string{
					name,
					city,
					coords,
				})
			}
		}

		playerList, _ := pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(list).Srender()
		area.Update(
			pterm.LightYellow(logger.Uptime()),
			"\n",
			pterm.LightBlue(pterm.Sprintf("Players Online: %d", playersOnline)),
			"\n",
			playerList,
		)

		time.Sleep(time.Second)
	}
}

func NewGame(db *sql.DB) *Game {
	p, _ := pterm.DefaultProgressbar.WithTotal(3).WithTitle("Generating Objects..").WithRemoveWhenDone().Start()

	p.UpdateTitle("Building Ranks..")
	LinkRanks()
	pterm.Success.Println(fmt.Sprintf("Generated Ranks: %d", len(RanksList)))
	p.Increment()

	p.UpdateTitle("Building Items..")
	GenerateItemsList()
	pterm.Success.Println(fmt.Sprintf("Generated Items: %d", len(ItemsList)))
	p.Increment()

	p.UpdateTitle("Building Cities..")
	cityList := GenerateCities()
	pterm.Success.Println(fmt.Sprintf("Generated Cities: %d", len(cityList)))
	p.Increment()

	game := Game{
		DbConn:       db,
		Clients:      make(map[*Client]bool),
		Players:      make(map[*Entity]bool),
		Register:     make(chan *Client),
		Logout:       make(chan *Client),
		GlobalEvents: make(chan protoreflect.ProtoMessage),
		NewsFlash:    make(chan protoreflect.ProtoMessage),
		World:        cityList,
	}

	p, _ = pterm.DefaultProgressbar.WithTotal(len(game.World)).WithTitle("Populating Cities..").WithRemoveWhenDone().Start()

	for _, city := range game.World {
		p.UpdateTitle(city.Name)
		city.Setup()
		city.StartCityTimers()
		city.Game = &game
		pterm.Success.Println("Done: " + city.Name)
		p.Increment()
	}

	logger.Logger.Info("Setup Complete")
	return &game
}
