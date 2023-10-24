package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/mreliasen/swi-server/internal/responses"
)

type ActionType uint8

const (
	CombatActionAim    = 1
	CombatActionPunch  = 2
	CombatActionStrike = 3
	CombatActionShoot  = 4
	CombatActionDeath  = 5
	CombatActionFlee   = 6
)

type CombatAction struct {
	Attacker  *Entity
	Target    *Entity
	Action    ActionType
	Direction *Coordinates
	Success   bool
}

func (c *CombatAction) SameLocation() bool {
	if !c.Attacker.Loc.Coords.SameAs(&c.Target.Loc.Coords) {
		if c.Attacker.CurrentTarget == c.Target {
			c.Attacker.CurrentTarget = nil
			delete(c.Target.TargetedBy, c.Attacker)
		}

		if c.Attacker.IsPlayer {
			c.Attacker.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_NORMAL,
				Messages: []string{"That target is not at the same location as you"},
			})
			return false
		}
	}

	return true
}

func (c *CombatAction) Execute() {
	if (c.Target == nil || c.Attacker == nil) && c.Action != CombatActionFlee {
		logger.Logger.Warn(fmt.Sprintf("%v tried to %v at %v", c.Attacker, c.Action, c.Target))
		return
	}

	if c.Action != CombatActionAim && c.Action != CombatActionFlee && c.Attacker.IsPlayer {
		now := time.Now().UnixMilli()
		until := c.Attacker.LastAttack + settings.PlayerAttackDelayMs

		if until > now {
			c.Attacker.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_NORMAL,
				Messages: []string{fmt.Sprintf("You must wait another %d seconds before you can attack again.", (until-now)/1000)},
			})
			return
		}

		c.Attacker.Mu.Lock()
		c.Attacker.LastAttack = now
		c.Attacker.Mu.Unlock()
	}

	switch c.Action {
	case CombatActionAim:
		c.aim()
	case CombatActionPunch:
		c.punch()
	case CombatActionShoot:
		c.shoot()
	case CombatActionStrike:
		c.strike()
	case CombatActionFlee:
		c.flee()
	}
}

func (c *CombatAction) flee() {
	dropItems := rand.Intn(3) + 1
	positions := []int{}

	for index, item := range c.Target.Inventory.Items {
		if item != nil {
			positions = append(positions, index)
		}
	}

	maxItems := len(positions)
	if dropItems > maxItems {
		dropItems = maxItems
	}

	if val, ok := c.Target.Loc.City.Grid[c.Direction.toString()]; ok {
		if maxItems >= dropItems {
			for i := 0; i < dropItems; i++ {
				pos := rand.Intn(maxItems)
				event := c.Target.Inventory.drop(positions[pos])
				c.Target.Loc.AddItem <- event
				logger.LogItems(c.Target.Name, "flee", event.Item.TemplateName, c.Target.Loc.Coords.North, c.Target.Loc.Coords.East, c.Target.Loc.Coords.City)
			}
		}

		val.PlayerJoin <- c.Target.Client
	}
}

func (c *CombatAction) aim() {
	if !c.SameLocation() {
		return
	}

	c.Attacker.Mu.Lock()
	c.Target.Mu.Lock()

	c.Attacker.CurrentTarget = c.Target
	c.Target.TargetedBy[c.Attacker] = true

	if !c.Target.IsPlayer {
		c.Target.NpcHostiles[c.Attacker.Client.UUID] = true
	}

	c.Target.Mu.Unlock()
	c.Attacker.Mu.Unlock()

	if c.Attacker.IsPlayer {
		c.Attacker.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
			Messages: []string{fmt.Sprintf("You take aim on %s", c.Target.Name)},
		})
	}

	if c.Target.IsPlayer {
		c.Target.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
			Messages: []string{fmt.Sprintf("%s takes aim on you.", c.Attacker.Name)},
		})
	}

	val := CreateEvent(map[uint64]bool{c.Attacker.PlayerID: true, c.Target.PlayerID: true}, &responses.Generic{
		Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
		Messages: []string{fmt.Sprintf("You see %s take aim on %s.", c.Attacker.Name, c.Target.Name)},
	})

	logger.LogCombat(c.Attacker.Name, c.Target.Name, "aim", "", 0, c.Attacker.Loc.Coords.North, c.Attacker.Loc.Coords.East, c.Attacker.Loc.City.ShortName)

	c.Target.Loc.Events <- &val
}

func (c *CombatAction) punch() {
	if !c.SameLocation() {
		return
	}

	c.Target.Mu.Lock()
	defer c.Target.Mu.Unlock()

	if c.Target.Dead {
		return
	}

	if c.Attacker.IsPlayer {
		c.Attacker.Mu.Lock()
		c.Attacker.AutoAttackType = CombatActionPunch
		c.Attacker.Mu.Unlock()
	}

	if !c.Attacker.SkillAcc.SkillCheck() {
		if c.Attacker.IsPlayer {
			c.Attacker.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_NORMAL,
				Messages: []string{fmt.Sprintf("You take a swing at %s but miss.", c.Target.Name)},
			})
		}

		if c.Target.IsPlayer {
			c.Target.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
				Messages: []string{fmt.Sprintf("%s takes a swing at you, but miss.", c.Attacker.Name)},
			})
		}

		val := CreateEvent(map[uint64]bool{c.Attacker.PlayerID: true, c.Target.PlayerID: true}, &responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
			Messages: []string{fmt.Sprintf("You see %s take a swing at %s but miss.", c.Attacker.Name, c.Target.Name)},
		})

		c.Target.Loc.Events <- &val
		return
	}

	c.Target.Health -= 2

	if c.Attacker.IsPlayer {
		c.Attacker.PlayerSendStatsUpdate()
		c.Attacker.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
			Messages: []string{fmt.Sprintf("You land a solid punch on %s.", c.Target.Name)},
		})
	}

	if c.Target.IsPlayer {
		c.Target.PlayerSendStatsUpdate()
		c.Target.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_ERROR,
			Messages: []string{fmt.Sprintf("%s lands a damaging blow on you.", c.Attacker.Name)},
		})
	}

	val := CreateEvent(map[uint64]bool{c.Attacker.PlayerID: true, c.Target.PlayerID: true}, &responses.Generic{
		Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
		Messages: []string{fmt.Sprintf("You see %s land a solid blow to %s head.", c.Attacker.Name, c.Target.Name)},
	})

	logger.LogCombat(c.Attacker.Name, c.Target.Name, "punch", "-", 2, c.Attacker.Loc.Coords.North, c.Attacker.Loc.Coords.East, c.Attacker.Loc.City.ShortName)

	c.Target.Loc.Events <- &val

	if c.Target.Health <= 0 {
		go c.Target.Death(c.Attacker)
		return
	} else {
		for client := range c.Attacker.Loc.Players {
			client.Player.sendGameFrame(false)
		}
	}
}

func (c *CombatAction) shoot() {
	if !c.SameLocation() {
		return
	}

	c.Target.Mu.Lock()
	defer c.Target.Mu.Unlock()

	if c.Target.Dead {
		return
	}

	if c.Attacker.IsPlayer {
		c.Attacker.Mu.Lock()
		c.Attacker.AutoAttackType = CombatActionShoot
		c.Attacker.Mu.Unlock()
	}

	weapon := c.Attacker.Inventory.Equipment[ItemTypeGun]
	if !c.Attacker.IsPlayer && weapon == nil {
		logger.Logger.Fatal(fmt.Sprintf("%s tried to shoot, but has no gun.", c.Attacker.NpcTitle))
	}

	if c.Attacker.IsPlayer && weapon == nil {
		c.Attacker.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
			Messages: []string{"You don't have a gun equipped"},
		})
		return
	}

	ammo := c.Attacker.Inventory.Equipment[ItemTypeAmmo]
	if c.Attacker.IsPlayer && ammo == nil {
		c.Attacker.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
			Messages: []string{"Click, click.. You are out of ammo."},
		})
		return
	}

	if c.Attacker.IsPlayer {
		if weapon.Condition <= 0 || rand.Float32() > weapon.Condition {
			c.Attacker.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
				Messages: []string{"Click, jam, silence.. your gun jammed! You clear the jam so you can attempt again."},
			})
			return
		}
	}

	if c.Attacker.IsPlayer {
		cond := 0.005 + ammo.GetAmmoWear()
		weapon.Condition -= cond
	}

	if !c.Attacker.SkillAcc.SkillCheck() {
		if c.Attacker.IsPlayer {
			c.Attacker.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_NORMAL,
				Messages: []string{fmt.Sprintf("You fire your %s at %s, but miss.", weapon.GetName(), c.Target.Name)},
			})
		}

		if c.Target.IsPlayer {
			c.Target.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
				Messages: []string{fmt.Sprintf("%s fires thier %s at you, but miss.", weapon.GetName(), c.Attacker.Name)},
			})
		}

		val := CreateEvent(map[uint64]bool{c.Attacker.PlayerID: true, c.Target.PlayerID: true}, &responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
			Messages: []string{fmt.Sprintf("You see %s fire their %s at %s, but miss.", c.Attacker.Name, weapon.GetName(), c.Target.Name)},
		})

		c.Target.Loc.Events <- &val
		return
	}

	dmg := weapon.GetDamage() + ammo.GetDamage()
	dmgReduction := uint(0)
	armor := c.Target.Inventory.Equipment[ItemTypeArmor]

	if armor != nil {
		dmgReduction = armor.GetArmorGuns()

		if c.Target.IsPlayer {
			armor.Condition -= float32(dmg) / 100
		}
	}

	dmg -= dmgReduction

	if dmg < 1 {
		dmg = 1
	}

	c.Target.Health -= int(dmg)

	if c.Attacker.IsPlayer {
		ammo.Amount -= 1

		if ammo.Amount <= 0 {
			for slot, itm := range c.Attacker.Inventory.Items {
				if itm == ammo {
					c.Attacker.Inventory.drop(slot)
					break
				}
			}
		}

		c.Attacker.PlayerSendStatsUpdate()
		c.Attacker.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
			Messages: []string{fmt.Sprintf("You fire you %s at %s, it's a direct hit!", weapon.GetName(), c.Target.Name)},
		})
	}

	if c.Target.IsPlayer {
		c.Target.PlayerSendStatsUpdate()
		c.Target.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_ERROR,
			Messages: []string{fmt.Sprintf("%s fires their %s at you, it's a direct hit.", c.Attacker.Name, weapon.GetName())},
		})
	}

	val := CreateEvent(map[uint64]bool{c.Attacker.PlayerID: true, c.Target.PlayerID: true}, &responses.Generic{
		Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
		Messages: []string{fmt.Sprintf("You see %s fire their %s at %s, landing a direct hit.", c.Attacker.Name, weapon.GetName(), c.Target.Name)},
	})

	c.Target.Loc.Events <- &val

	logger.LogCombat(c.Attacker.Name, c.Target.Name, "shoot", weapon.GetName(), dmg, c.Attacker.Loc.Coords.North, c.Attacker.Loc.Coords.East, c.Attacker.Loc.City.ShortName)

	if c.Target.IsPlayer && armor != nil {
		if armor.Condition <= 0 {
			for slot, item := range c.Target.Inventory.Items {
				if item == armor {
					c.Target.Inventory.drop(slot)
					break
				}
			}

			c.Target.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
				Messages: []string{fmt.Sprintf("Your %s breaks, the pieces fall to the ground.", armor.GetName())},
			})

			c.Target.PlayerSendInventoryUpdate()
		}
	}

	if c.Target.Health <= 0 {
		go c.Target.Death(c.Attacker)
		return
	} else {
		for client := range c.Attacker.Loc.Players {
			client.Player.sendGameFrame(false)
		}
	}
}

func (c *CombatAction) strike() {
	if !c.SameLocation() {
		return
	}

	c.Target.Mu.Lock()
	defer c.Target.Mu.Unlock()

	if c.Target.Dead {
		return
	}

	if c.Attacker.IsPlayer {
		c.Attacker.Mu.Lock()
		c.Attacker.AutoAttackType = CombatActionStrike
		c.Attacker.Mu.Unlock()
	}

	weapon := c.Attacker.Inventory.Equipment[ItemTypeMelee]
	if !c.Attacker.IsPlayer && weapon == nil {
		logger.Logger.Fatal(fmt.Sprintf("%s tried to strike, but has no melee.", c.Attacker.NpcTitle))
	}

	if c.Attacker.IsPlayer && weapon == nil {
		c.Attacker.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
			Messages: []string{"You don't have a melee weapon equipped"},
		})
		return
	}

	if !c.Attacker.SkillAcc.SkillCheck() {
		if c.Attacker.IsPlayer {
			c.Attacker.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_NORMAL,
				Messages: []string{fmt.Sprintf("You strike at %s with your %s, but miss.", c.Target.Name, weapon.GetName())},
			})
		}

		if c.Target.IsPlayer {
			c.Target.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
				Messages: []string{fmt.Sprintf("%s lunges at you with a %s, but miss.", c.Attacker.Name, weapon.GetName())},
			})
		}

		val := CreateEvent(map[uint64]bool{c.Attacker.PlayerID: true, c.Target.PlayerID: true}, &responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
			Messages: []string{fmt.Sprintf("You see %s strike at %s with a %s, but miss.", c.Attacker.Name, c.Target.Name, weapon.GetName())},
		})

		c.Target.Loc.Events <- &val
		return
	}

	dmg := weapon.GetDamage()
	dmgReduction := uint(0)
	armor := c.Target.Inventory.Equipment[ItemTypeArmor]

	if armor != nil {
		dmgReduction = armor.GetArmorMelee()

		if c.Target.IsPlayer {
			armor.Condition -= float32(dmg) / 100
		}
	}

	if c.Attacker.IsPlayer {
		cond := 0.005 + float32(dmgReduction)/100
		weapon.Condition -= cond
	}

	dmg -= dmgReduction

	if dmg < 1 {
		dmg = 1
	}

	c.Target.Health -= int(dmg)

	if c.Attacker.IsPlayer {
		c.Attacker.PlayerSendStatsUpdate()
		c.Attacker.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_SUCCESS,
			Messages: []string{fmt.Sprintf("You land a solid strike on %s with your %s", c.Target.Name, weapon.GetName())},
		})
	}

	if c.Target.IsPlayer {
		c.Target.PlayerSendStatsUpdate()
		c.Target.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_ERROR,
			Messages: []string{fmt.Sprintf("%s lands a solid strike on you with a %s.", c.Attacker.Name, weapon.GetName())},
		})
	}

	val := CreateEvent(map[uint64]bool{c.Attacker.PlayerID: true, c.Target.PlayerID: true}, &responses.Generic{
		Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
		Messages: []string{fmt.Sprintf("You see %s land a solid strike on %s with a %s.", c.Attacker.Name, c.Target.Name, weapon.GetName())},
	})

	c.Target.Loc.Events <- &val

	logger.LogCombat(c.Attacker.Name, c.Target.Name, "strike", weapon.GetName(), dmg, c.Attacker.Loc.Coords.North, c.Attacker.Loc.Coords.East, c.Attacker.Loc.City.ShortName)

	if c.Target.IsPlayer && armor != nil {
		if armor.Condition <= 0 {
			for slot, item := range c.Target.Inventory.Items {
				if item == armor {
					c.Target.Inventory.drop(slot)
					break
				}
			}

			c.Target.PlayerSendStatsUpdate()
			c.Target.Client.SendEvent(&responses.Generic{
				Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
				Messages: []string{fmt.Sprintf("Your %s breaks, the pieces fall to the ground.", armor.GetName())},
			})

			c.Target.PlayerSendInventoryUpdate()
		}
	}

	if c.Target.Health <= 0 {
		go c.Target.Death(c.Attacker)
		return
	} else {
		for client := range c.Attacker.Loc.Players {
			client.Player.sendGameFrame(false)
		}
	}
}
