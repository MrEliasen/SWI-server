package game

import (
	"errors"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/game/skills"
	"github.com/mreliasen/swi-server/internal/database/models"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/mreliasen/swi-server/internal/responses"
)

type PlayerGameFrame struct {
	Id      uint64 `json:"id"`
	Name    string `json:"name"`
	Rank    string `json:"rank"`
	GangTag string `json:"gang_tag"`
}

func (p *Entity) PlayerGameFrame() *responses.Player {
	return &responses.Player{
		Id:      p.PlayerID,
		Name:    p.Name,
		Rank:    p.Rank.Name,
		GangTag: p.GangTag(),
	}
}

func (p *Entity) GangTag() string {
	if p.Gang == nil {
		return ""
	}

	return p.Gang.Tag
}

func (p *Entity) PlayerGetBuildingCommand(cmdKey string) (*Command, bool) {
	if p.Loc != nil {
		if val, ok := p.Loc.getBuildingCommand(cmdKey); ok {
			return val, true
		}
	}

	return nil, false
}

func (p *Entity) PlayerSendMapUpdate() {
	hasItem, _ := p.Inventory.HasItem("smartphone")

	event := responses.GPS{
		Enabled: hasItem,
	}

	if hasItem {
		pois := []*responses.Coordinate{}
		for _, poi := range p.Loc.City.POILocations {
			pois = append(pois, poi.toResponse())
		}

		event.Name = p.Loc.City.Name
		event.Height = int32(p.Loc.City.Height)
		event.Width = int32(p.Loc.City.Width)
		event.Pois = pois
	}

	p.Client.SendEvent(&event)
}

func (p *Entity) PlayerSendLocationUpdate() {
	hasItem, _ := p.Inventory.HasItem("smartphone")

	event := responses.GPS{
		Enabled:  hasItem,
		Location: p.Loc.Coords.toResponse(),
	}

	p.Client.SendEvent(&event)
}

func (p *Entity) PlayerSendPlayerList() {
	for player := range p.Client.Game.Players {
		event := &responses.PlayerList{
			Type:     responses.PlayerEvent_EVENT_TYPE_PLAYER_JOIN,
			Id:       player.Client.UUID,
			Name:     player.Name,
			Hometown: player.Hometown,
			GangTag:  player.GangTag(),
		}

		p.Client.Send <- event
	}
}

func (p *Entity) PlayerSendStatsUpdate() {
	event := responses.Stats{
		Name:       p.Name,
		Cash:       p.Cash,
		Bank:       p.Bank,
		Rank:       p.Rank.Name,
		Reputation: p.Reputation,
		NextRank:   0,
		Hometown:   p.Hometown,
		Health:     int32(p.Health),
		MaxHealth:  settings.PlayerMaxHealth,
		Skills:     []*responses.Skill{},
	}

	if p.Rank.NextRank != nil {
		event.NextRank = p.Rank.NextRank.MinRep - p.Rank.MinRep
	}

	event.Skills = append(event.Skills, &responses.Skill{
		Key:   "Accuracy",
		Value: p.SkillAcc.Value,
	})
	/* event.Skills = append(event.Skills, KeyFloatValue{
		Key:   "Hide",
		Value: p.SkillHide.Value,
	})
	event.Skills = append(event.Skills, KeyFloatValue{
		Key:   "Snoop",
		Value: p.SkillSnoop.Value,
	})
	event.Skills = append(event.Skills, KeyFloatValue{
		Key:   "Search",
		Value: p.SkillSearch.Value,
	})
	event.Skills = append(event.Skills, KeyFloatValue{
		Key:   "Track",
		Value: p.SkillTrack.Value,
	}) */

	p.Client.SendEvent(&event)
}

func (p *Entity) PlayerSendInventoryUpdate() {
	p.Client.SendEvent(p.Inventory.GameFrame())
}

func (p *Entity) sendGameFrame(clearFrame bool) {
	if p.Loc == nil {
		logger.Logger.Trace("Can't send game frame, no location")
		return
	}

	players := []*responses.Player{}
	for client := range p.Loc.Players {
		if client.Player.PlayerID == p.PlayerID {
			continue
		}

		players = append(players, client.Player.PlayerGameFrame())
	}

	npcs := []*responses.NPC{}
	for npc := range p.Loc.Npcs {
		npcs = append(npcs, npc.NPCGameFrame())
	}

	frame := responses.Location{
		Clear:       clearFrame,
		CityName:    p.Loc.City.Name,
		Description: p.Loc.Description,
		Coordinates: p.Loc.Coords.toResponse(),
		Buildings:   p.Loc.buildingGameFrames(),
		Players:     players,
		Npcs:        npcs,
		Items:       p.Loc.itemsGameFrames(),
	}

	p.Client.SendEvent(&frame)
}

func CharacterToPlayer(c *models.Character) (*Entity, *Coordinates, error) {
	if c.Id == 0 {
		return nil, nil, errors.New("invalid character")
	}

	p := Entity{
		UserId:            c.UserId,
		PlayerID:          c.Id,
		Name:              c.Name,
		Bank:              c.Bank,
		Cash:              c.Cash,
		Health:            int(c.Health),
		Reputation:        c.Reputation,
		Hometown:          c.Hometown,
		NpcKills:          c.NpcKills,
		PlayerKills:       c.PlayerKills,
		IsAdmin:           c.IsAdmin == 1,
		IsPlayer:          true,
		SkillAcc:          skills.Accuracy{Value: c.SkillAcc},
		SkillHide:         skills.Hide{Value: c.SkillHide},
		SkillSearch:       skills.Search{Value: c.SkillSearch},
		SkillSnoop:        skills.Snoop{Value: c.SkillSnoop},
		SkillTrack:        skills.Track{Value: c.SkillTrack},
		LastMove:          time.Now().UnixMilli(),
		ShoppingWith:      make(map[*Entity]int64),
		TargetedBy:        make(map[*Entity]bool),
		AutoAttackEnabled: false,
		AutoAttackType:    CombatActionPunch,
	}

	p.Inventory = NewInventory(&p)

	lastLocation := Coordinates{
		North: c.LocationNorth,
		East:  c.LocationEast,
		City:  c.LocationCity,
	}

	return &p, &lastLocation, nil
}

func (e *Entity) StartRoutines() {
	go func() {
		for {
			if !e.AutoAttackEnabled {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if e.CurrentTarget == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if e.Loc == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if e.Client == nil || e.Client.Connection == nil || e.Client.CombatLogging {
				return
			}

			now := time.Now().UnixMilli()
			until := e.LastAttack + settings.PlayerAttackDelayMs

			if until > now {
				time.Sleep(time.Duration(until-now) * time.Millisecond)
			}

			if e.CurrentTarget == nil {
				continue
			}

			action := CombatAction{
				Target:   e.CurrentTarget,
				Attacker: e,
				Action:   e.AutoAttackType,
			}
			action.Execute()
		}
	}()
}

func (e *Entity) Save() {
	if !e.IsPlayer {
		return
	}

	e.Mu.Lock()

	_, err := e.Client.Game.DbConn.Exec(
		`UPDATE
            characters
        SET 
            reputation = ?,
            health = ?,
            npc_kill = ?,
            player_kills = ?,
            cash = ?,
            bank = ?,
            skill_acc = ?,
            skill_hide = ?,
            skill_track = ?,
            skill_snoop = ?,
            skill_search = ?,
            location_n = ?,
            location_e = ?,
            location_city = ?
        WHERE
            user_id = ?`,
		e.Reputation,
		e.Health,
		e.NpcKills,
		e.PlayerKills,
		e.Cash,
		e.Bank,
		e.SkillAcc.Value,
		e.SkillHide.Value,
		e.SkillTrack.Value,
		e.SkillSnoop.Value,
		e.SkillSearch.Value,
		e.LastLocation.North,
		e.LastLocation.East,
		e.LastLocation.City,
		e.Client.UserId,
	)
	if err != nil {
		print(err.Error())
	}

	e.Mu.Unlock()
	e.Inventory.save()
}
