package game

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal"
	"github.com/mreliasen/swi-server/internal/database/models"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/mreliasen/swi-server/internal/responses"
)

var CommandAliases = map[string]string{
	"/get": "/pickup",
	"/p":   "/pickup",
	"/w":   "/pm",
	"/h":   "/help",
	"/s":   "/say",
	"/g":   "/global",
	"/m":   "/move",
	"/r":   "/refresh",
	"/b":   "/buy",
}

var CommandsList = map[string]*Command{
	/* "/track": {
			Args:        []string{"player-name"},
			Description: "Try to track the location of a player.",
			AllowInGame: true,
			Help: func(c *Client) {
			},
			Call: func(c *Client, args []string) {
				if c.Player.Loc == nil {
					return
				}

				playername := strings.ToLower(args[0])

				if !c.Player.SkillTrack.SkillCheck() {
					c.SendEvent(&responses.Generic{
											Messages: []string{"You failed to track down the location of this player."},
					})
					return
				}

				var player *Player
				for p := range c.Game.Players {
					if strings.ToLower(p.Name) == playername {
						player = p
					}
				}

				chance := 0.0
	            if c.Player.SkillTrack.Value < 35{
	            }

				c.SendEvent(&responses.Generic{
					Messages: []string{"%s was last seen in %s"},
				})
			},
		}, */
	"/use": {
		Args:        []string{"ID"},
		Description: "Uses a given item by ID",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(args) != 1 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid command "},
				})
				return
			}

			c.Player.Inventory.useById(args[0])
		},
	},
	"/shopsell": {
		Args:        []string{"merchantID", "inventoryIndex"},
		Description: "Sells the item at the given inventory index to the given merchant(id)",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(args) < 2 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid command, try /shop <id> <inventory index>"},
				})
				return
			}

			buildingId := args[0]
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid sell index."},
				})
				return
			}

			var building *Building
			for _, b := range c.Player.Loc.Buildings {
				if b.ID == buildingId {
					building = b
				}
			}

			if building == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid building."},
				})
				return
			}

			item := c.Player.Inventory.Items[index]
			if item == nil {
				return
			}

			building.Mu.Lock()
			defer building.Mu.Unlock()

			amount, ok := building.ShopBuyType[item.GetItemType()]
			if !ok {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "I am not interested in that..",
				})
				return
			}

			if amount == 0 {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "I am not looking to buy more of that type of item.",
				})
				return
			}

			itemEvent := c.Player.Inventory.drop(int(index))

			if itemEvent == nil {
				return
			}

			money := uint32(float32(itemEvent.Item.GetPrice()) * settings.ItemSellPriceLoss)

			if money == 0 {
				money = 1
			}

			c.Player.Mu.Lock()
			c.Player.Cash += money
			c.Player.Mu.Unlock()

			if amount != -1 {
				building.ShopBuyType[itemEvent.Item.GetItemType()] -= 1
			}

			logger.LogBuySell(c.Player.Name, "sell", money, itemEvent.Item.TemplateName)

			building.SyncShopInventory(c)
			c.Player.PlayerSendInventoryUpdate()
			c.Player.PlayerSendStatsUpdate()

			c.SendEvent(&responses.MerchantMessage{
				Status:  responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
				Message: fmt.Sprintf("Here's $%d for the %s", money, itemEvent.Item.GetName()),
			})
		},
	},
	"/shopbuy": {
		Args:        []string{"shop type", "inventory index"},
		Description: "Buys the item at the given inventory index from the given shop",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) < 2 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid command, try /shop <id> <inventory index>"},
				})
				return
			}

			buildingId := args[0]
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid sell index."},
				})
				return
			}

			var building *Building
			for _, b := range c.Player.Loc.Buildings {
				if b.ID == buildingId {
					building = b
				}
			}

			if building == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid building."},
				})
				return
			}

			if building.ShopStock == nil || len(building.ShopStock) <= 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid shop."},
				})
				return
			}

			if index < 0 || int(index) >= len(building.ShopStock) {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid index."},
				})
				return
			}

			itemToBuy := building.ShopStock[index]

			if itemToBuy.Amount == 0 {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "I have none of those left, pick something else.",
				})
				return
			}

			itemTemplate := ItemsList[itemToBuy.TemplateId]

			if itemTemplate.GetMinRep() > c.Player.Reputation {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "I don't know you well enough to sell you that. Come back when your name is more known.",
				})
				return
			}

			c.Player.Mu.Lock()
			defer c.Player.Mu.Unlock()

			if itemTemplate.GetPrice() > c.Player.Cash {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "You do not have enough money on you, come back when you have cash.",
				})
				return
			}

			if !c.Player.Inventory.HasRoom() {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "You don't have enough room to buy that.",
				})
				return
			}

			newItem, ok := NewItem(itemToBuy.TemplateId)
			if !ok {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "I can't seem to find that item.. odd. Come back later (Error)",
				})
				return
			}

			c.Player.Cash -= itemTemplate.GetPrice()
			err = c.Player.Inventory.addItem(newItem)
			if err != nil {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "You don't seem to have room, I've dropped the item on the ground (Inventory error)",
				})

				c.Player.Loc.AddItem <- &ItemMoved{
					Item: newItem,
					By:   building.Name,
				}
				return
			}

			logger.LogBuySell(c.Player.Name, "buy", itemTemplate.GetPrice(), itemTemplate.TemplateName)

			c.SendEvent(&responses.MerchantMessage{
				Status:  responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
				Message: fmt.Sprintf("Done. Here's your %s.", newItem.GetName()),
			})

			building.SyncShopInventory(c)
			c.Player.PlayerSendInventoryUpdate()
			c.Player.PlayerSendStatsUpdate()
		},
	},
	"/selldrug": {
		Args:        []string{"merchant id", "inventory index"},
		Description: "Sells the item at the given inventory index to the given merchant(id)",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player == nil {
				return
			}

			if len(args) < 2 {
				return
			}

			var druggie *Entity
			sellerID := args[0]
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return
			}

			for npc := range c.Player.ShoppingWith {
				if npc.NpcType == DrugAddict && sellerID == npc.NpcID {
					druggie = npc
					break
				}
			}

			if druggie == nil {
				c.SendEvent(&responses.Generic{
					Status:   responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{"There is no one here to tell to."},
				})
				return
			}

			if isHostile := druggie.NpcHostiles[c.UUID]; isHostile {
				return
			}

			item := c.Player.Inventory.Items[index]
			if item == nil {
				return
			}

			if item.GetItemType() != ItemTypeDrug {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "I am not interested in that..",
				})
				return
			}

			if !druggie.Inventory.HasRoom() {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "I got what I need, go sell your shit to someone else.",
				})
				return
			}

			itemEvent := c.Player.Inventory.drop(int(index))

			if itemEvent == nil {
				return
			}

			price := uint32((float32(itemEvent.Item.GetPrice()) * settings.DrugProfitMargin) * c.Player.Loc.City.DrugDemands[item.TemplateName])

			if price <= 0 {
				price = 1
			}

			c.Player.Mu.Lock()
			c.Player.Cash += price
			c.Player.Reputation += settings.DrugRepIncrease
			c.Player.Mu.Unlock()

			logger.LogBuySell(c.Player.Name, "sell", price, itemEvent.Item.TemplateName)

			druggie.Inventory.addItem(itemEvent.Item)
			druggie.SyncDruggieTrade(c)
			c.Player.PlayerSendInventoryUpdate()
			c.Player.PlayerSendStatsUpdate()

			c.SendEvent(&responses.MerchantMessage{
				Status:  responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
				Message: fmt.Sprintf("Here's $%d for the %s", price, itemEvent.Item.GetName()),
			})
		},
	},
	"/purchase": {
		Args:        []string{"merchant id", "inventory index"},
		Description: "Buys the item at the given inventory index from the given merchant(id)",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player == nil {
				return
			}

			if len(args) < 2 {
				return
			}

			var dealer *Entity
			sellerID := args[0]
			index, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return
			}

			for t := range c.Player.Loc.Npcs {
				if t.NpcType == DrugDealer && t.NpcID == sellerID {
					dealer = t
					break
				}
			}

			if isHostile := dealer.NpcHostiles[c.UUID]; isHostile {
				return
			}

			dealer.Inventory.Mu.Lock()
			c.Player.Inventory.Mu.Lock()
			defer dealer.Inventory.Mu.Unlock()
			defer c.Player.Inventory.Mu.Unlock()

			if !c.Player.Inventory.HasRoom() {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "There is no room left in your inventory.",
				})
				return
			}

			item := dealer.Inventory.Items[index]
			if item == nil {
				return
			}

			c.Player.Mu.Lock()

			price := uint32(float32(item.GetPrice()) * c.Player.Loc.City.DrugDemands[item.TemplateName])

			if price <= 0 {
				price = 1
			}

			if c.Player.Cash < price {
				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Message: "You do not have enough cash for that.",
				})
				c.Player.Mu.Unlock()
				return
			}

			logger.LogBuySell(c.Player.Name, "buy", price, item.TemplateName)
			c.Player.Mu.Unlock()

			go func() {
				c.Player.Mu.Lock()
				defer c.Player.Mu.Unlock()

				itemEvent := dealer.Inventory.drop(int(index))

				if itemEvent == nil {
					return
				}

				c.Player.Cash -= price
				c.Player.Inventory.addItem(itemEvent.Item)
				dealer.SyncDealerInventory(c)
				c.Player.PlayerSendInventoryUpdate()
				c.Player.PlayerSendStatsUpdate()

				c.SendEvent(&responses.MerchantMessage{
					Status:  responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
					Message: fmt.Sprintf("Here is your %s, anything else?", itemEvent.Item.GetName()),
				})
			}()
		},
	},
	"/closetrade": {
		Args:        []string{"name"},
		Description: "close active trade",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			if c.Player == nil {
				return
			}

			c.Player.Mu.Lock()

			for t := range c.Player.ShoppingWith {
				t.Mu.Lock()
				delete(t.ShoppingWith, c.Player)
				delete(c.Player.ShoppingWith, t)
				t.Mu.Unlock()
			}

			c.Player.Mu.Unlock()
		},
	},
	"/sell": {
		Args:        []string{"name"},
		Description: "Opens drug sellng menu",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player == nil {
				return
			}

			var druggie *Entity
			var altDruggie *Entity
			name := ""

			if len(args) > 0 {
				name = strings.ToLower(args[0])
			}

			for t := range c.Player.Loc.Npcs {
				if t.NpcType == DrugAddict {
					altDruggie = t

					if strings.HasPrefix(strings.ToLower(t.Name), name) {
						druggie = t
						break
					}
				}
			}

			if druggie == nil {
				druggie = altDruggie

				if druggie == nil {
					c.SendEvent(&responses.Generic{
						Messages: []string{"There are no druggies by that name."},
					})
					return
				}
			}

			druggie.OpenDruggieMenu(c)
		},
	},
	"/buy": {
		Args:        []string{"name"},
		Description: "Opens drug dealers buy menu",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player == nil {
				return
			}

			var dealer *Entity
			var altDealer *Entity
			name := ""

			if len(args) > 0 {
				name = strings.ToLower(args[0])
			}

			for t := range c.Player.Loc.Npcs {
				if t.NpcType == DrugDealer {
					altDealer = t

					if strings.HasPrefix(strings.ToLower(t.Name), name) {
						dealer = t
						break
					}
				}
			}

			if dealer == nil {
				dealer = altDealer

				if dealer == nil {
					c.SendEvent(&responses.Generic{
						Messages: []string{"There are no drug dealers here going by that name."},
					})
					return
				}
			}

			dealer.OpenDrugDealer(c)
		},
	},
	"/unequip": {
		Args:        []string{"itemID"},
		Description: "Unequips the given item",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(args) != 1 {
				return
			}

			index, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || index < 0 {
				return
			}

			c.Player.Inventory.unequip(int(index))
			c.Player.PlayerSendInventoryUpdate()
			c.Player.PlayerSendStatsUpdate()
		},
	},
	"/info": {
		Args:        []string{"itemID"},
		Description: "Gives info on the item",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(args) != 1 {
				return
			}

			ID := args[0]

			c.Player.Inventory.Info(ID)
		},
	},
	"/equip": {
		Args:        []string{"itemID"},
		Description: "Equips the given item",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(args) != 1 {
				return
			}

			ID := args[0]

			c.Player.Inventory.equip(ID)
			c.Player.PlayerSendInventoryUpdate()
			c.Player.PlayerSendStatsUpdate()
		},
	},
	"/unaim": {
		Args:        []string{},
		Description: "Removes your target lock",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			if c.Player.Loc == nil {
				return
			}

			if c.Player.CurrentTarget == nil {
				return
			}

			c.SendEvent(&responses.Generic{
				Messages: []string{fmt.Sprintf("You stop taking aim at %s", c.Player.CurrentTarget.Name)},
			})

			c.Player.CurrentTarget.Client.SendEvent(&responses.Generic{
				Messages: []string{fmt.Sprintf("%s stops taking aim at you.", c.Player.Name)},
			})

			delete(c.Player.CurrentTarget.TargetedBy, c.Player)
			c.Player.CurrentTarget = nil
		},
	},
	"/aim": {
		Args:        []string{"target"},
		Description: "Sets your for your attacks. The target cannot move away while aim'ed at",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(args) == 0 {
				return
			}

			var target *Entity
			name := strings.ToLower(args[0])

			for t := range c.Player.Loc.Players {
				if strings.HasPrefix(strings.ToLower(t.Player.Name), name) {
					target = t.Player
					break
				}
			}

			if target == nil {
				for t := range c.Player.Loc.Npcs {
					if strings.HasPrefix(strings.ToLower(t.Name), name) {
						target = t
						break
					}
				}
			}

			if target == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"There are no one here going by that name."},
				})
				return
			}

			action := CombatAction{
				Target:   target,
				Attacker: c.Player,
				Action:   CombatActionAim,
			}
			action.Execute()
		},
	},
	"/punch": {
		Args:        []string{},
		Description: "Throws a punch at the target",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			if c.Player.Loc == nil {
				return
			}

			if c.Player.CurrentTarget == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You have no target, so you throw some punches into the air."},
				})
				return
			}

			action := CombatAction{
				Target:   c.Player.CurrentTarget,
				Attacker: c.Player,
				Action:   CombatActionPunch,
			}
			action.Execute()
		},
	},
	"/strike": {
		Args:        []string{},
		Description: "Strikes the target with your melee weapon",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			if c.Player.Loc == nil {
				return
			}

			if c.Player.CurrentTarget == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You have no target."},
				})
				return
			}

			action := CombatAction{
				Target:   c.Player.CurrentTarget,
				Attacker: c.Player,
				Action:   CombatActionStrike,
			}
			action.Execute()
		},
	},
	"/shoot": {
		Args:        []string{},
		Description: "Shoots at the target with your gun",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			if c.Player.Loc == nil {
				return
			}

			if c.Player.CurrentTarget == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You have no target."},
				})
				return
			}

			action := CombatAction{
				Target:   c.Player.CurrentTarget,
				Attacker: c.Player,
				Action:   CombatActionShoot,
			}
			action.Execute()
		},
	},
	"/autoattack": {
		Args:        []string{},
		Description: "Will toggle auto attack on/off.",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			c.Player.Mu.Lock()
			c.Player.AutoAttackEnabled = !c.Player.AutoAttackEnabled
			c.Player.Mu.Unlock()

			c.SendEvent(&responses.Generic{
				Status: responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{
					"Auto Attack, when enabled, will attack an /aim'ed target with the last type of attack used (default: Punch)",
					fmt.Sprintf("Auto Attack enabled: %v", c.Player.AutoAttackEnabled),
				},
			})
		},
	},
	"/flee": {
		Args:        []string{"direction"},
		Description: "Run away from a fight, in the choosen direction, but you could drop some items in the process",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(c.Player.TargetedBy) == 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"No one is targeting you, no need to flee"},
				})
				return
			}

			north := c.Player.Loc.Coords.North
			east := c.Player.Loc.Coords.East

			switch strings.ToLower(args[0]) {
			case "up", "w", "north":
				north += 1
			case "down", "s", "south":
				north += -1
			case "left", "a", "west":
				east += -1
			case "right", "d", "east":
				east += 1
			}

			if north < 0 || north > int(c.Player.Loc.City.Height) || east < 0 || east > int(c.Player.Loc.City.Width) {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_WARN,
					Messages: []string{
						"You cannot move any further in that direction that direction",
					},
				})
				return
			}

			action := CombatAction{
				Target: c.Player,
				Action: CombatActionFlee,
				Direction: &Coordinates{
					North: c.Player.Loc.Coords.North + north,
					East:  c.Player.Loc.Coords.East + east,
				},
			}
			action.Execute()
		},
	},
	"/pickup": {
		Args:        []string{"itemid or name"},
		Description: "Pickup item of a given name or specific id",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(c.Player.Loc.Items) == 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"There are no items laying on the ground."},
				})
				return
			}

			var puItem *Item

			if len(args) > 0 {
				nameArg := strings.ToLower(args[0])
				idArg := args[0]
				var itmById *Item
				var itmByName *Item

				for item := range c.Player.Loc.Items {
					if item.ID == idArg {
						itmById = item
						itmByName = nil
						break
					}

					if strings.HasPrefix(strings.ToLower(item.GetName()), nameArg) {
						itmByName = item
						break
					}
				}

				if itmById != nil {
					puItem = itmById
				}
				if itmByName != nil {
					puItem = itmByName
				}

				if puItem == nil {
					c.SendEvent(&responses.Generic{
						Messages: []string{"There are no items on the ground beginning with that."},
					})
					return
				}
			} else {
				for item := range c.Player.Loc.Items {
					puItem = item
					break
				}
			}

			logger.LogItems(c.Player.Name, "pickup", puItem.TemplateName, c.Player.Loc.Coords.North, c.Player.Loc.Coords.East, c.Player.Loc.Coords.City)

			c.Player.Loc.RemoveItem <- &ItemMoved{
				Item:   puItem,
				By:     c.Player.Name,
				Player: c.Player,
			}
		},
	},
	"/drop": {
		Args:        []string{"inventory slot"},
		Description: "Drop an item at a given inventory slot",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				return
			}

			if c.Player.Loc == nil {
				return
			}

			slotIndex, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			event := c.Player.Inventory.drop(int(slotIndex))
			if event != nil {
				c.Player.Loc.AddItem <- event
			}

			logger.LogItems(c.Player.Name, "drop", event.Item.TemplateName, c.Player.Loc.Coords.North, c.Player.Loc.Coords.East, c.Player.Loc.Coords.City)
		},
	},
	"/npcs": {
		Args:          []string{},
		Description:   "List of all NPCs.",
		AllowInGame:   true,
		AllowUnAuthed: true,
		AllowAuthed:   true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			headings := []string{"Name", "health", "Accuracy", "Damage", "Armor (Range)", "Armor (Melee)"}
			data := [][]string{}

			for _, tmpl := range NpcTemplates {
				dmg := 2
				drRange := 0
				drMelee := 0

				if tmpl.Equipment != nil {
					if gunId, ok := tmpl.Equipment[ItemTypeGun]; ok {
						gun := ItemsList[gunId]

						if gun != nil {
							dmg = int(gun.Damage)
						}

						if ammoId, ok := tmpl.Equipment[ItemTypeAmmo]; ok {
							ammo := ItemsList[ammoId]

							if ammo != nil {
								dmg += int(ammo.Damage)
							}
						}
					}

					if meleeId, ok := tmpl.Equipment[ItemTypeMelee]; ok {
						weapon := ItemsList[meleeId]

						if weapon != nil {
							dmg = int(weapon.Damage)
						}
					}

					if armorId, ok := tmpl.Equipment[ItemTypeArmor]; ok {
						armor := ItemsList[armorId]

						if armor != nil {
							drMelee = int(armor.ArmorGuns)
							drRange = int(armor.ArmorMelee)
						}
					}
				}

				data = append(data, []string{
					tmpl.Title,
					fmt.Sprintf("%d", tmpl.Health),
					fmt.Sprintf("%.2f", tmpl.SkillAcc),
					fmt.Sprintf("%d", dmg),
					fmt.Sprintf("%d", drRange),
					fmt.Sprintf("%d", drMelee),
				})
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, data),
			})
		},
	},
	"/help": {
		Args:          []string{},
		Description:   "List of all general commands and controls.",
		AllowInGame:   true,
		AllowUnAuthed: true,
		AllowAuthed:   true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: []string{" ", " ", "Here is a list of all common commands:"},
			})

			headings := []string{"Command", "Description"}
			commands := [][]string{
				{"/<command> help", "Most commands have a help menu, explaining its use."},
				{"/here", "Will tell you what commands are available at a given location (if any)"},
				{"/say", "Send a chat message visible only to those in the same location."},
				{"/global", "Send a chat message visible to all online players."},
				{"/pm <name> <msg>", "Send a private chat message to the specified player."},
				{"/buy <name>", "Open drug dealer buy menu"},
				{"/sell <name>", "Open drug addict sell menu"},
				{"/shop", "Open the shop window (eg. arms dealer)"},
				{"/aim <name>", "Takes aim at a target, required before you attack"},
				{"/unaim", "Removes your lock on the target"},
				{"/flee <dir>", "Escapes from battle, but you loose items."},
				{"/autoattack", "Toggles auto-attack on/off."},
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, commands),
			})

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: []string{" ", " ", "This is all the controls:"},
			})

			ctrlHeadings := []string{"Keyboard Key", "Alt. Key", "Description"}
			ctrls := [][]string{
				{"W", "K, Arrow Up", "Move 1 space north"},
				{"A", "H, Arrow Left", "Move 1 space West"},
				{"S", "J, Arrow Down", "Move 1 space South"},
				{"D", "L, Arrow Right", "Move 1 space East"},
				{"R", "", "Clears event log, and gets the location latest changes"},
				{"I", "Enter", "Focuses the command input if not already focused."},
				{"Esc.", "", "Unfocuses the command input, if in focus"},
				{"1", "", "Strike your target with your fist (2dmg)"},
				{"2", "", "Strike your target with your melee weapon (weapon dmg)"},
				{"3", "", "Shoot your target with you gun (weapon dmg)"},
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(ctrlHeadings, ctrls),
			})

			c.SendEvent(&responses.Generic{
				Messages: []string{
					"Deal drug to gain money and increase your reputation. The more reputation you have, the more things you can buy and do.",
					"This game is full-loot PvPvE, anything player or NPC carry will be dropped when killed, and any cash on them will go to the killer.",
					"",
				},
			})
		},
	},
	"/here": {
		Args:        []string{},
		Description: "Gives you information about available commands at a given location",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			if c.Player.Loc == nil {
				return
			}

			if len(c.Player.Loc.Buildings) == 0 {
				return
			}

			headings := []string{"Command", "Description", "Example"}
			commands := [][]string{}
			for _, b := range c.Player.Loc.Buildings {
				for cmd, c := range b.Commands {
					commands = append(commands, []string{
						cmd + " " + strings.Join(c.Args, " "),
						c.Description,
						c.Example,
					})
				}
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, commands),
			})
		},
	},
	"/pm": {
		Args:        []string{"user", "message"},
		Description: "Send a private message to a player",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) != 2 {
				return
			}

			playerName := args[0]
			message := strings.Join(args[1:], " ")
			var player *Entity
			for tc := range c.Game.Clients {
				if tc.Player == nil {
					return
				}

				if strings.ToLower(tc.Player.Name) == playerName {
					player = tc.Player
					break
				}
			}

			if player == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"There are no one online going by that name."},
				})
				return
			}

			logger.LogChat("private", c.Player.Name, message, player.Name)

			msg := []byte(message)
			msg = bytes.TrimSpace(bytes.ReplaceAll(msg, []byte{'\n'}, []byte{' '}))

			player.Client.SendEvent(&responses.Chat{
				Type:   responses.ChatType_CHAT_TYPE_PRIVATE,
				Player: c.Player.PlayerGameFrame(),
				Msg:    string(msg),
			})
		},
	},
	"/say": {
		Args:        []string{"message"},
		Description: "Send a message locally",
		AllowInGame: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				return
			}

			if c.Player.Loc == nil {
				return
			}

			message := strings.Join(args, " ")
			logger.LogChat("local", c.Player.Name, message, "")

			msg := []byte(message)
			msg = bytes.TrimSpace(bytes.ReplaceAll(msg, []byte{'\n'}, []byte{' '}))

			event := responses.Generic{
				Status: responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{
					fmt.Sprintf("(Local) %s: %s", c.Player.Name, string(msg)),
				},
			}

			c.Player.Loc.Events <- &ClientResponse{
				Payload: &event,
			}
		},
	},
	"/global": {
		Args:          []string{"message"},
		Description:   "Send a message to the global chat.",
		AllowInGame:   true,
		AllowAuthed:   false,
		AllowUnAuthed: false,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				return
			}

			message := strings.Join(args, " ")

			logger.LogChat("global", c.Player.Name, message, "")

			msg := []byte(message)
			msg = bytes.TrimSpace(bytes.ReplaceAll(msg, []byte{'\n'}, []byte{' '}))

			event := responses.Chat{
				Type:   responses.ChatType_CHAT_TYPE_GLOBAL,
				Player: c.Player.PlayerGameFrame(),
				Msg:    string(msg),
			}

			c.Game.GlobalEvents <- &event
		},
	},
	"/refresh": {
		Args:          []string{},
		Description:   "Refresh the game frame",
		AllowInGame:   true,
		AllowAuthed:   false,
		AllowUnAuthed: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			if c.Player == nil {
				c.Game.SendMOTD(c)
				return
			}

			c.Player.sendGameFrame(true)
		},
	},
	"/move": {
		Args:          []string{"direction"},
		Description:   "Move in a provided direction.",
		AllowInGame:   true,
		AllowAuthed:   false,
		AllowUnAuthed: false,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				return
			}

			if c.Player.Loc == nil {
				return
			}

			if len(c.Player.TargetedBy) > 0 {
				names := []string{}
				for p := range c.Player.TargetedBy {
					names = append(names, p.Name)
				}

				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						fmt.Sprintf("You cannot move while being held up by: %s", strings.Join(names, ",")),
						"You can flee the area using: \"/flee <direction>\", eg \"/free north\", but you will drop some items.",
					},
				})
				return
			}

			now := time.Now().UnixMilli()
			if c.Player.LastMove+settings.PlayerMoveDelayMs > now {
				return
			}
			c.Player.LastMove = now

			north := c.Player.Loc.Coords.North
			east := c.Player.Loc.Coords.East

			switch strings.ToLower(args[0]) {
			case "up", "w", "north":
				north += 1
			case "down", "s", "south":
				north += -1
			case "left", "a", "west":
				east += -1
			case "right", "d", "east":
				east += 1
			}

			if north < 0 || north > int(c.Player.Loc.City.Height) || east < 0 || east > int(c.Player.Loc.City.Width) {
				c.SendEvent(&responses.Generic{
					Messages: []string{
						"You cannot go any further in that direction",
					},
				})
				return
			}

			newLocation := Coordinates{
				North: int(north),
				East:  int(east),
			}

			if val, ok := c.Player.Loc.City.Grid[newLocation.toString()]; ok {
				val.PlayerJoin <- c
			}
		},
	},
	"/new": {
		Args:          []string{"name", "city"},
		Description:   "Creates a new character",
		AllowInGame:   false,
		AllowAuthed:   true,
		AllowUnAuthed: false,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			_, _, err := c.Game.GetUserCharacter(c.UserId)
			if err == nil {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"You cannot crate a character. You are not logged in",
					},
				})
				return
			}

			if len(args) < 2 {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Invalid command. The format is:  \"/login username password\"",
					},
				})
				return
			}

			name := args[0]
			city := args[1]

			if _, ok := c.Game.World[city]; !ok {
				list := []string{}

				for c := range c.Game.World {
					list = append(list, c)
				}

				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Invalid City. Your options are:",
						strings.Join(list, ", "),
					},
				})
				return
			}

			row := c.Game.DbConn.QueryRow("SELECT id FROM characters WHERE LOWER(name) = ?", strings.ToLower(name))

			char := models.Character{}
			row.Scan(&char.Id)

			if char.Id != 0 {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"That name is already taken.",
					},
				})
				return
			}

			hometown := c.Game.World[city]
			startLocation := hometown.RandomLocation()

			newChar := models.Character{
				UserId:        c.UserId,
				Name:          name,
				Cash:          settings.PlayerStartCash,
				Reputation:    settings.PlayerStartReputation,
				Bank:          settings.PlayerStartBank,
				Health:        settings.PlayerMaxHealth,
				SkillAcc:      settings.PlayerStartSkillAcc,
				SkillTrack:    settings.PlayerStartSkillTrack,
				SkillHide:     settings.PlayerStartSkillHide,
				SkillSnoop:    settings.PlayerStartSkillSnoop,
				SkillSearch:   settings.PlayerStartSkillSearch,
				Hometown:      hometown.ShortName,
				LocationNorth: startLocation.North,
				LocationEast:  startLocation.East,
				LocationCity:  hometown.ShortName,
				CreatedAt:     time.Now().Unix(),
			}

			result, err := c.Game.DbConn.Exec(
				"INSERT INTO characters (user_id, name, reputation, health, npc_kill, player_kills, cash, bank, hometown, skill_acc, skill_track, skill_hide, skill_snoop, skill_search, location_n, location_e, location_city, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
				&newChar.UserId,
				&newChar.Name,
				&newChar.Reputation,
				&newChar.Health,
				0,
				0,
				&newChar.Cash,
				&newChar.Bank,
				&newChar.Hometown,
				&newChar.SkillAcc,
				&newChar.SkillTrack,
				&newChar.SkillHide,
				&newChar.SkillSnoop,
				&newChar.SkillSearch,
				&newChar.LocationNorth,
				&newChar.LocationEast,
				&newChar.LocationCity,
				&newChar.CreatedAt,
			)
			if err != nil {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Failed to create character, system error.",
					},
				})
				return
			}

			lastId, err := result.LastInsertId()
			if err != nil {
				log.Print(err)
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Failed to create character, system error.",
					},
				})
				return
			}

			newChar.Id = uint64(lastId)
			player, lastLocation, err := c.Game.GetUserCharacter(newChar.UserId)
			if err != nil {
				log.Print(err)
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Failed load in your character, system error.",
					},
				})
				return
			}

			c.Game.LoginPlayer(c, player, lastLocation)
		},
	},
	"/authenticate": {
		Args:          []string{"username", "password"},
		Description:   "Login to your account.",
		AllowInGame:   false,
		AllowAuthed:   false,
		AllowUnAuthed: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) < 2 {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Invalid command. The format is:  \"/login username password\"",
					},
				})
				return
			}
			email := args[0]
			password := args[1]

			row := c.Game.DbConn.QueryRow("SELECT id, password, user_type FROM users WHERE email = ? LIMIT 1", email)
			if row == nil {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Invalid username and password combination.",
					},
				})
				return
			}

			user := models.User{}
			err := row.Scan(&user.Id, &user.Password, &user.UserType)
			if user.Id == 0 || err != nil {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Invalid username and password combination.",
					},
				})
				return
			}

			if ok, err := internal.CheckPassword(password, user.Password); !ok || err != nil {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_ERROR,
					Messages: []string{
						"Invalid username and password combination.",
					},
				})
				return
			}

			c.Authenticated = true
			c.UserId = user.Id
			c.UserType = uint8(user.UserType)

			player, lastLocation, err := c.Game.GetUserCharacter(user.Id)
			if err != nil {
				c.SendEvent(&responses.Generic{
					Status: responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
					Messages: []string{
						"Your have been logged in, however you do not have a character yet.",
						"To create a new character type /new",
					},
				})
				return
			}

			c.Game.LoginPlayer(c, player, lastLocation)
		},
	},
	"/demand": {
		Args:         []string{},
		Description:  "update drug demand for the city",
		AllowInGame:  true,
		AdminCommand: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			c.Player.Loc.City.UpdateDrugDemand()
		},
	},
	"/save": {
		Args:         []string{},
		Description:  "Save all online players",
		AllowInGame:  true,
		AdminCommand: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			c.Game.Save()
		},
	},
	"/restock": {
		Args:         []string{},
		Description:  "Restocks all drugs on the server",
		AllowInGame:  true,
		AdminCommand: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, _ []string) {
			c.Game.Restock()
		},
	},
	"/additem": {
		Args:         []string{"itemid"},
		Description:  "Spawn a new item",
		AllowInGame:  true,
		AdminCommand: true,
		Help: func(c *Client) {
		},
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				return
			}

			if c.Player.Loc == nil {
				return
			}

			item, ok := NewItem(args[0])
			if !ok {
				return
			}

			c.Player.Loc.AddItem <- &ItemMoved{
				Item: item,
				By:   c.Player.Name,
			}
		},
	},
}
