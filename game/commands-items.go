package game

import (
	"fmt"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal"
	"github.com/mreliasen/swi-server/internal/responses"
)

type ItemUseEffect struct {
	Use func(*Client, *Item, int)
}

var UseEffectsList = map[string]*ItemUseEffect{
	"smartphone": {
		Use: func(c *Client, _ *Item, _ int) {
			if c.Player.Loc == nil {
				return
			}

			c.Player.Mu.Lock()
			defer c.Player.Mu.Unlock()

			var useCost int64 = settings.SmartPhoneCost

			if c.Player.Bank < useCost {
				c.SendEvent(&responses.Generic{
					Messages: []string{fmt.Sprintf("This information does not come for free, you don't have the $%d it costs in your bank.", useCost)},
				})
			}

			c.Player.Bank -= useCost

			ctrlHeadings := []string{"Who", "Last Known Location"}
			locations := [][]string{}

			for npc := range c.Player.Loc.City.NPCs[DrugDealer] {
				if npc.Loc == nil {
					continue
				}

				locations = append(locations, []string{
					fmt.Sprintf("%s the %s", npc.Name, npc.NpcTitle),
					fmt.Sprintf("N%d-E%d", npc.Loc.Coords.North, npc.Loc.Coords.East),
				})
			}

			for npc := range c.Player.Loc.City.NPCs[DrugAddict] {
				if npc.Loc == nil {
					continue
				}

				locations = append(locations, []string{
					fmt.Sprintf("%s the %s", npc.Name, npc.NpcTitle),
					fmt.Sprintf("N%d-E%d", npc.Loc.Coords.North, npc.Loc.Coords.East),
				})
			}

			c.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{fmt.Sprintf("Informant: \"(Phone) Your $%d was received. Alright, here is what I know..\"", useCost)},
			})

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(ctrlHeadings, locations),
			})

			c.Player.PlayerSendStatsUpdate()
		},
	},
	"usedrug": {
		Use: func(c *Client, item *Item, slotIndex int) {
			c.Player.Mu.Lock()
			c.Player.Inventory.Mu.Lock()

			healthCost := settings.DrugUseHealthCost
			var repGain int64 = settings.DrugUseRepGain

			if c.Player.Health <= healthCost {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Using this would kill you.."},
				})

				c.Player.Mu.Unlock()
				c.Player.Inventory.Mu.Unlock()
				return
			}

			c.Player.Health -= healthCost
			c.Player.Reputation += repGain
			c.Player.Mu.Unlock()

			item.Amount -= 1
			c.Player.Inventory.Mu.Unlock()

			if item.Amount <= 0 {
				c.Player.Inventory.drop(slotIndex)
			}

			c.SendEvent(&responses.Generic{
				Messages: []string{fmt.Sprintf("You use the %s. It didn't do your health any favours. (-%d Health, +%d Rep)", item.GetName(), healthCost, repGain)},
			})

			c.Player.PlayerSendInventoryUpdate()
			c.Player.PlayerSendStatsUpdate()
		},
	},
}
