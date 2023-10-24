package game

import (
	"math/rand"
	"sync"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal/responses"
)

type City struct {
	Name              string
	ShortName         string
	ISO               string
	Grid              map[string]*Location
	Width             uint
	Height            uint
	TravelCostMin     uint32
	TravelCostMax     uint32
	TravelCost        uint32
	Players           map[*Client]bool
	PlayerJoin        chan *Client
	PlayerLeave       chan *Client
	CityEvents        chan bool
	Game              *Game
	NPCs              map[NPCType]map[*Entity]bool
	NpcSpawnList      map[NPCType]uint8
	BuildingLocations []BuildingLocation
	POILocations      []Coordinates
	DrugDemands       map[string]float32
	Mu                sync.Mutex
}

type BuildingLocation struct {
	Coords    Coordinates
	Buildings []BuildingType
}

func (c *City) RandomiseTravelCost() {
	c.Mu.Lock()
	res := rand.Intn(int(c.TravelCostMax)) + int(c.TravelCostMin)
	c.TravelCost = uint32(res)
	c.Mu.Unlock()
}

func (c *City) StartCityTimers() {
	go func() {
		for {
			nextRefresh := rand.Intn(settings.CityDemandUpdateMaxMins) + settings.CityDemandUpdateMinMins
			time.Sleep(time.Minute * time.Duration(nextRefresh))
			c.UpdateDrugDemand()
		}
	}()

	go func() {
		for {
			client := <-c.PlayerJoin
			c.Mu.Lock()
			c.Players[client] = true
			c.Mu.Unlock()
		}
	}()

	go func() {
		for {
			client := <-c.PlayerLeave
			c.Mu.Lock()
			delete(c.Players, client)
			c.Mu.Unlock()
		}
	}()

	go func() {
		for {
			<-c.CityEvents
			c.Mu.Lock()

			event := &responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{"Informant: \"(Phone) Yo, the demand for different dope has changed. If you need any directions just call me\""},
			}

			for client := range c.Players {
				if ok, _ := client.Player.Inventory.HasItem("smartphone"); ok {
					client.SendEvent(event)
				}
			}

			c.Mu.Unlock()
		}
	}()
}

func (c *City) UpdateDrugDemand() {
	c.Mu.Lock()
	demand := map[string]float32{}

	for _, t := range DrugsList {
		v := float32(rand.NormFloat64() / 3)

		if v < 0 {
			v *= -1.0
		} else {
			v += 1.0
		}

		demand[t.TemplateName] = v
	}

	c.DrugDemands = demand
	c.Mu.Unlock()
	c.CityEvents <- true
}

func (c *City) GeneratePOIs() {
	pois := []Coordinates{}

	for _, poi := range c.BuildingLocations {
		for _, bType := range poi.Buildings {
			pois = append(pois, Coordinates{
				North:   poi.Coords.North,
				East:    poi.Coords.East,
				POI:     BuildingTemplates[bType].Name,
				POIType: bType,
			})
		}
	}

	c.POILocations = pois
}

func (c *City) RandomLocation() Coordinates {
	north := rand.Intn(int(c.Height))
	east := rand.Intn(int(c.Width))
	return Coordinates{
		North: north,
		East:  east,
	}
}

func (c *City) Setup() {
	for n := 0; n <= int(c.Height); n++ {
		for e := 0; e <= int(c.Height); e++ {
			loc := CreateLocation(c, n, e)
			c.Grid[loc.Coords.toString()] = loc
		}
	}

	// spawn NPCs
	for npcType, amount := range c.NpcSpawnList {
		c.NPCs[npcType] = make(map[*Entity]bool)

		for i := uint8(0); i < amount; i++ {
			npc := NewNPC(npcType)
			c.NPCs[npcType][npc] = true
			coords := c.RandomLocation()
			c.Grid[coords.toString()].NpcJoin <- npc
		}
	}

	c.GeneratePOIs()
}
