package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/mreliasen/swi-server/internal/responses"
)

var BuildingCommandsList = map[string]*Command{
	"/shop": {
		Args:        []string{"name"},
		Description: "Opens the shop's trade menu",
		AllowInGame: true,
		Help: func(c *Client) {
			headings := []string{"Example", "Description"}
			lines := [][]string{
				{"/shop", "Will open the first found shop in the area."},
				{"/shop <name>", "Will open the first sho beginning with the name."},
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, lines),
			})
		},
		Call: func(c *Client, args []string) {
			if c.Player == nil {
				return
			}

			var building *Building
			var altBuilding *Building
			name := ""

			if len(args) > 0 {
				name = strings.ToLower(args[0])
			}

			for _, b := range c.Player.Loc.Buildings {
				if b.ShopStock != nil && len(b.ShopStock) > 0 || b.ShopBuyType != nil && len(b.ShopBuyType) > 0 {
					altBuilding = b

					if strings.HasPrefix(strings.ToLower(b.Name), name) {
						building = b
						break
					}
				}
			}

			if building == nil {
				building = altBuilding

				if building == nil {
					c.SendEvent(&responses.Generic{
						Messages: []string{"There are no shops around here by that name"},
					})
					return
				}
			}

			building.OpenShop(c)
		},
	},
	"/drink": {
		Args:        []string{"amount"},
		Description: "Drink and buy rounds to increase you reputation.",
		Example:     "/drink 10",
		AllowInGame: true,
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Missing amount. Try: /drink help"},
				})
				return
			}

			amount, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || amount < 1 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid amount. Try: /drink help"},
				})
				return
			}

			health_cost := int(amount * settings.DrinkHealthCost)
			drink_cost := int64(amount * settings.DrinkCost)
			rep_gain := int64(amount * settings.DrinkRepGain)

			if c.Player.Cash < drink_cost {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You do not have enough money on you"},
				})
				return
			}

			if c.Player.Health <= health_cost {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Don't be stupid, that will kill you."},
				})
				return
			}

			c.Player.Cash -= drink_cost
			c.Player.Health -= health_cost
			c.Player.Reputation += rep_gain

			c.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{fmt.Sprintf("You spend %d on buying drinks and paying for strippers, your reputation increased by %d", drink_cost, rep_gain)},
			})

			go c.Player.PlayerSendStatsUpdate()
		},
		Help: func(c *Client) {
			headings := []string{"Example", "Health Cost", "$ Cost", "Rep Gain"}
			lines := [][]string{
				{"/drink 1", fmt.Sprintf("%d", settings.DrinkHealthCost), fmt.Sprintf("%d", settings.DrinkCost), fmt.Sprintf("%d", settings.DrinkRepGain)},
				{"/drink 5", fmt.Sprintf("%d", 5*settings.DrinkHealthCost), fmt.Sprintf("%d", 5*settings.DrinkCost), fmt.Sprintf("%d", 5*settings.DrinkRepGain)},
				{"/drink 10", fmt.Sprintf("%d", 10*settings.DrinkHealthCost), fmt.Sprintf("%d", 10*settings.DrinkCost), fmt.Sprintf("%d", 10*settings.DrinkRepGain)},
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, lines),
			})
		},
	},
	"/transfer": {
		Args:        []string{"user", "amount"},
		Description: "Transfer money from your bank account to another.",
		Example:     "/transfer <User> 100",
		AllowInGame: true,
		Call: func(c *Client, args []string) {
			if c.Player.Hometown != c.Player.Loc.City.ShortName {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You can only use the bank to withdraw money when you are not in your home city."},
				})
				return
			}

			username := args[0]
			amount, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil || amount < 1 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid amount. Try: /transfer <username> <amount>"},
				})
				return
			}

			c.Player.Mu.Lock()
			defer c.Player.Mu.Unlock()

			if amount > c.Player.Bank {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You don't have that much money in your bank account."},
				})
				return
			}

			var player *Entity
			for tc := range c.Game.Clients {
				if tc.Player == nil {
					return
				}

				if strings.ToLower(tc.Player.Name) == username {
					player = tc.Player
					break
				}
			}

			if player == nil {
				c.SendEvent(&responses.Generic{
					Messages: []string{"We cannot find anyone going by that name."},
				})
				return
			}

			player.Mu.Lock()
			defer player.Mu.Unlock()

			c.Player.Bank -= amount
			player.Bank += amount

			c.SendEvent(&responses.Generic{
				Messages: []string{fmt.Sprintf("You just transferred $%d to %s", amount, player.Name)},
			})

			player.Client.SendEvent(&responses.Generic{
				Messages: []string{fmt.Sprintf("%s just transferred you $%d", c.Player.Name, amount)},
			})

			logger.LogMoney(c.Player.Name, "transfer", amount, player.Name)
		},
	},
	"/deposit": {
		Args:        []string{"amount"},
		Description: "Deposit money in your bank account.",
		Example:     "/deposit 100",
		AllowInGame: true,
		Call: func(c *Client, args []string) {
			if c.Player.Hometown != c.Player.Loc.City.ShortName {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You can only use the bank to withdraw money when you are not in your home city."},
				})
				return
			}

			if len(args) == 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Missing amount. Try: /deposit help"},
				})
				return
			}

			amount, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || amount < 1 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid amount. Try: /deposit help"},
				})
				return
			}

			if c.Player.Cash < amount {
				c.SendEvent(&responses.Generic{
					Messages: []string{fmt.Sprintf("You do not have %d on you", amount)},
				})
				return
			}

			c.Player.Cash -= amount
			c.Player.Bank += amount

			c.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{fmt.Sprintf("You deposit $%d in your bank account.", amount)},
			})

			logger.LogMoney(c.Player.Name, "deposit", amount, "")

			go c.Player.PlayerSendStatsUpdate()
		},
		Help: func(c *Client) {
			headings := []string{"Example", ""}
			lines := [][]string{
				{"/deposit 100", "Deposits $100"},
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, lines),
			})
		},
	},
	"/withdraw": {
		Args:        []string{"amount"},
		Description: "Withdraw money from your bank account.",
		Example:     "/withdraw 100",
		AllowInGame: true,
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Missing amount. Try: /withdraw help"},
				})
				return
			}

			argAmount, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || argAmount < 1 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid amount. Try: /withdraw help"},
				})
				return
			}

			amount := argAmount

			if c.Player.Bank < amount {
				c.SendEvent(&responses.Generic{
					Messages: []string{fmt.Sprintf("You do not have $%d in the bank", amount)},
				})
				return
			}

			c.Player.Bank -= amount
			c.Player.Cash += amount

			c.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{fmt.Sprintf("You withdraw $%d from your account.", amount)},
			})

			logger.LogMoney(c.Player.Name, "withdraw", amount, "")

			go c.Player.PlayerSendStatsUpdate()
		},
		Help: func(c *Client) {
			headings := []string{"Example", ""}
			lines := [][]string{
				{"/withdraw 100", "Withdraws $100"},
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, lines),
			})
		},
	},
	"/heal": {
		Args:        []string{"amount"},
		Description: "You can heal up here.",
		Example:     "/heal 12",
		AllowInGame: true,
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Missing amount. Try: /heal help"},
				})
				return
			}

			if c.Player.Health >= settings.PlayerMaxHealth {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You do not need any patching up, you are at full health"},
				})
				return
			}

			amount, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || amount < 1 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid amount. Try: /heal help"},
				})
				return
			}

			healAmount := int(amount)

			if healAmount > settings.PlayerMaxHealth-c.Player.Health {
				healAmount = settings.PlayerMaxHealth - c.Player.Health
			}

			healCost := int64(healAmount * settings.HealCostPerPoint)

			if c.Player.Cash < healCost {
				c.SendEvent(&responses.Generic{
					Messages: []string{fmt.Sprintf("You do not have enough money. Costs %d per point to heal", settings.HealCostPerPoint)},
				})
				return
			}

			c.Player.Cash -= healCost
			c.Player.Health += healAmount

			if c.Player.Health > settings.PlayerMaxHealth {
				c.Player.Health = settings.PlayerMaxHealth
			}

			c.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
				Messages: []string{fmt.Sprintf("You pay the doctor %d to patch you up.", healAmount*settings.HealCostPerPoint)},
			})

			go c.Player.PlayerSendStatsUpdate()
		},
		Help: func(c *Client) {
			headings := []string{"Example", "Cost per HP", "Total Cost"}
			lines := [][]string{
				{"/heal 10", fmt.Sprintf("%d", settings.HealCostPerPoint), fmt.Sprintf("%d", 10*settings.HealCostPerPoint)},
				{"/heal 43", fmt.Sprintf("%d", settings.HealCostPerPoint), fmt.Sprintf("%d", 43*settings.HealCostPerPoint)},
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, lines),
			})
		},
	},
	"/travel": {
		Args:        []string{"destination"},
		Description: "travel to another city. /travel help for destinations",
		Example:     "/travel lon",
		AllowInGame: true,
		Call: func(c *Client, args []string) {
			if len(args) == 0 {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Missing destination. Try: /travel help"},
				})
				return
			}

			city, ok := c.Game.World[strings.ToUpper(args[0])]
			if !ok {
				c.SendEvent(&responses.Generic{
					Messages: []string{"Invalid destination. Try: /travel help"},
				})
				return
			}

			if c.Player.Cash < city.TravelCost {
				c.SendEvent(&responses.Generic{
					Messages: []string{"You do not have enough cash on you"},
				})
				return
			}

			c.Player.Cash -= city.TravelCost

			for _, poi := range city.POILocations {
				if poi.POIType == BuildingTypeAirport {
					city.Grid[poi.toString()].PlayerJoin <- c

					c.SendEvent(&responses.Generic{
						Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
						Messages: []string{"You fly off to your destination"},
					})

					go c.Player.PlayerSendStatsUpdate()
				}
			}
		},
		Help: func(c *Client) {
			headings := []string{"Destination", "Cost", "Command"}
			lines := [][]string{}

			for _, city := range c.Game.World {
				lines = append(lines, []string{
					city.Name,
					fmt.Sprintf("%d", city.TravelCost),
					fmt.Sprintf("/travel %s", city.ShortName),
				})
			}

			c.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable(headings, lines),
			})
		},
	},
}
