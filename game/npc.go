package game

import (
	"math/rand"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal/responses"
)

type (
	Gender  int8
	NPCType int16
)

const (
	GenderRandom Gender = 0
	GenderMale   Gender = 1
	GenderFemale Gender = 2
)

type NPCGameFrame struct {
	Name    string  `json:"name"`
	NpcType NPCType `json:"type"`
	Rank    string  `json:"rank"`
	Health  int     `json:"health"`
	Gender  Gender  `json:"gender"`
}

func (n *Entity) NPCGameFrame() *responses.NPC {
	return &responses.NPC{
		Name:   n.Name,
		Rank:   n.NpcTitle,
		Health: int32(n.Health),
	}
}

func (n *Entity) NPCStartRoutines() {
	go func() {
		for {
			time.Sleep(30 * time.Second)

			if len(n.ShoppingWith) == 0 {
				continue
			}

			n.Mu.Lock()
			for t := range n.ShoppingWith {
				if t.Loc.Coords.SameAs(&n.Loc.Coords) {
					continue
				}

				t.Mu.Lock()
				delete(t.ShoppingWith, n)
				delete(n.ShoppingWith, t)
				t.Mu.Unlock()
			}
			n.Mu.Unlock()
		}
	}()

	go func() {
		for {
			delay := settings.NPCMoveMinDelaySeconds + rand.Intn(settings.NPCMoveMaxDelaySeconds-settings.NPCMoveMinDelaySeconds)
			time.Sleep(time.Duration(delay) * time.Second)

			if n.Loc != nil {
				n.NPCMove()
			}
		}
	}()

	go func() {
		for {
			time.Sleep(settings.NPCAttackDelayMs * time.Millisecond)

			if n.Loc != nil {
				n.NPCAttack()
			}
		}
	}()
}

func (n *Entity) NPCFindTarget() (*Entity, bool) {
	if len(n.NpcHostiles) == 0 {
		return nil, false
	}

	if n.CurrentTarget != nil && n.CurrentTarget.IsPlayer {
		if _, ok := n.Loc.Players[n.CurrentTarget.Client]; ok {
			return n.CurrentTarget, true
		}
	}

	for p := range n.Loc.Players {
		if _, ok := n.NpcHostiles[p.Player.Client.UUID]; ok {
			return p.Player, true
		}
	}

	return nil, false
}

func (n *Entity) NPCAttack() {
	if n.CurrentTarget == nil {
		player, ok := n.NPCFindTarget()
		if !ok {
			return
		}

		action := CombatAction{
			Attacker: n,
			Target:   player,
			Action:   CombatActionAim,
		}
		action.Execute()
		return
	}

	var attackType ActionType = CombatActionPunch

	if n.Inventory.Equipment[ItemTypeMelee] != nil {
		attackType = CombatActionStrike
	}

	if n.Inventory.Equipment[ItemTypeGun] != nil {
		attackType = CombatActionShoot
	}

	action := CombatAction{
		Attacker: n,
		Target:   n.CurrentTarget,
		Action:   attackType,
	}
	action.Execute()
}

func (n *Entity) NPCMove() {
	if n.Loc == nil {
		return
	}

	if len(n.ShoppingWith) > 0 {
		return
	}

	if n.CurrentTarget != nil {
		return
	}

	north := n.Loc.Coords.North
	east := n.Loc.Coords.East

	rand_number := rand.Intn(4)
	switch rand_number {
	case 0:
		north += 1
	case 1:
		north += -1
	case 2:
		east += -1
	case 3:
		east += 1
	}

	if north < 0 {
		north += 2
	}

	if north > int(n.Loc.City.Height) {
		north -= 2
	}

	if east < 0 {
		east += 2
	}

	if east > int(n.Loc.City.Width) {
		east -= 2
	}

	newLocation := Coordinates{
		North: north,
		East:  east,
	}

	if val, ok := n.Loc.City.Grid[newLocation.toString()]; ok {
		val.NpcJoin <- n
	}
}

func (e *Entity) RandomiseGenderName() {
	template := NpcTemplates[e.NpcType]

	if template.Gender == GenderRandom {
		gender := GenderFemale

		if rand.Intn(2) == 1 {
			gender = GenderMale
		}

		e.NpcGender = gender
	}

	if e.NpcGender == GenderMale {
		e.Name = MaleNames[rand.Intn(len(MaleNames))]
	} else {
		e.Name = FemaleNames[rand.Intn(len(MaleNames))]
	}
}
