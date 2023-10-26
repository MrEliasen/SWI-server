package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/game/skills"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/mreliasen/swi-server/internal/responses"
)

type Entity struct {
	// Shared -----
	Mu            sync.Mutex
	Loc           *Location
	CurrentTarget *Entity
	TargetedBy    map[*Entity]bool
	Inventory     *Inventory
	Name          string
	Cash          uint32
	Health        int
	Dead          bool
	SkillAcc      skills.Accuracy
	IsPlayer      bool
	LastLocation  Coordinates
	// Player -----
	PlayerID          uint64
	Bank              uint32
	Gang              *Gang
	Hometown          string
	LastMove          int64
	LastSkill         map[string]int64
	LastAttack        int64
	PlayerKills       uint
	NpcKills          uint
	Reputation        int64
	UserId            uint64
	Rank              *Rank
	Client            *Client
	IsAdmin           bool
	AutoAttackEnabled bool
	AutoAttackType    ActionType
	SkillHide         skills.Hide
	SkillSearch       skills.Search
	SkillSnoop        skills.Snoop
	SkillTrack        skills.Track
	// NPC -----
	NpcID         string
	NpcCashReward uint32
	NpcGender     Gender
	NpcType       NPCType
	NpcTitle      string
	NpcRepReward  int32
	NpcCommands   map[string]*Command
	NpcHostiles   map[string]bool
	// Shopping -----
	ShoppingWith map[*Entity]int64
}

func (n *Entity) RemoveTargetLock() {
	if n.CurrentTarget != nil {
		delete(n.CurrentTarget.TargetedBy, n)
	}

	// clear all target locks
	n.CurrentTarget = nil

	if len(n.TargetedBy) > 0 {
		for t := range n.TargetedBy {
			t.CurrentTarget = nil
			// remove the attack from the victim target list
			delete(n.TargetedBy, t)
			// remove the victim from the attacker target list
			delete(t.TargetedBy, n)
		}
	}
}

func (n *Entity) Death(killer *Entity) {
	n.Dead = true

	// drop all items
	n.Inventory.dump(n.Loc)

	n.RemoveTargetLock()

	n.Mu.Lock()
	killer.Mu.Lock()

	if killer.IsPlayer {
		if n.IsPlayer {
			killer.PlayerKills += 1
		} else {
			killer.NpcKills += 1
		}
	}

	if n.IsPlayer {
		amount := n.Cash
		killer.Cash += n.Cash
		n.Cash = 50
		n.Health = 50
		n.SkillAcc.Value *= 0.96
		n.SkillHide.Value *= 0.96
		n.SkillSnoop.Value *= 0.96
		n.SkillTrack.Value *= 0.96
		n.SkillSearch.Value *= 0.96

		logger.LogMoney(n.Name, "death", amount, killer.Name)
	} else {
		killer.Cash += n.NpcCashReward
	}

	if !n.IsPlayer {
		killer.Reputation += int64(n.NpcRepReward)
	}

	killer.Mu.Unlock()
	n.Mu.Unlock()

	if n.IsPlayer {
		logger.LogCombat(killer.Name, n.Name, "kill", "", 0, n.Loc.Coords.North, n.Loc.Coords.East, n.Loc.City.Name)
	} else {
		logger.LogCombat(killer.Name, n.NpcTitle, "kill", "", 0, n.Loc.Coords.North, n.Loc.Coords.East, n.Loc.City.Name)
	}

	messageKiller := fmt.Sprintf("You put %s in their place, and of their items drop to the ground. They had $%d on them, which is now yours.", n.Name, n.Cash)
	messageWitness := fmt.Sprintf("You just witnessed %s put %s in their place, they lie bleeding out on the ground. All of %s items drop on the ground.", killer.Name, n.Name, n.Name)

	if !n.IsPlayer {
		messageKiller = fmt.Sprintf("%s drops dead, and all of their items drop on the ground.", n.Name)
		messageWitness = fmt.Sprintf("You just witnessed %s murder %s in cold blood. All of %sitems drop on the ground.", killer.Name, n.Name, n.Name)
	}

	if killer.IsPlayer {
		killer.Client.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
			Messages: []string{messageKiller},
		})
	}

	event := CreateEvent(map[uint64]bool{n.PlayerID: true, killer.PlayerID: true}, &responses.Generic{
		Status:   responses.ResponseStatus_RESPONSE_STATUS_INFO,
		Messages: []string{messageWitness},
	})

	n.Loc.Events <- &event

	if n.IsPlayer {
		for _, poi := range n.Loc.City.POILocations {
			if poi.POIType == BuildingTypeHospital {
				n.Loc.City.Grid[poi.toString()].Respawn <- CombatAction{
					Target:   n,
					Attacker: killer,
					Action:   CombatActionDeath,
				}
			}
		}

		n.Save()
		return
	}

	go func() {
		city := n.Loc.City
		delete(n.Loc.Npcs, n)
		n.Loc = nil

		for {
			time.Sleep(settings.NpcRespawnDelaySeconds * time.Second)
			npc := NewNPC(n.NpcType)
			delete(city.NPCs[n.NpcType], n)

			city.NPCs[npc.NpcType][npc] = true
			coords := city.RandomLocation()
			city.Grid[coords.toString()].NpcJoin <- npc
		}
	}()
}
