package game

import (
	"errors"

	"github.com/mreliasen/swi-server/internal/database/models"
	"github.com/mreliasen/swi-server/internal/logger"
)

func (g *Game) GetUserCharacter(userId uint64) (*Entity, *Coordinates, error) {
	row := g.DbConn.QueryRow(`
        SELECT
            id,
            name,
            reputation,
            health,
            npc_kill,
            player_kills,
            cash,
            bank,
            hometown,
            skill_acc,
            skill_track,
            skill_hide,
            skill_snoop,
            skill_search,
            location_n,
            location_e,
            location_city,
            gang_id,
            is_admin
        FROM
            characters
        WHERE
            user_id = ?
        LIMIT
            1
        `, userId)

	if row == nil {
		return nil, nil, errors.New("no character found")
	}

	character := models.Character{}
	err := row.Scan(
		&character.Id,
		&character.Name,
		&character.Reputation,
		&character.Health,
		&character.NpcKills,
		&character.PlayerKills,
		&character.Cash,
		&character.Bank,
		&character.Hometown,
		&character.SkillAcc,
		&character.SkillHide,
		&character.SkillSearch,
		&character.SkillTrack,
		&character.SkillSnoop,
		&character.LocationNorth,
		&character.LocationEast,
		&character.LocationCity,
		&character.GangId,
		&character.IsAdmin,
	)

	if err != nil || character.Id == 0 {
		logger.Logger.Error(err.Error())
		return nil, nil, errors.New("no character found")
	}

	player, lastLocation, err := CharacterToPlayer(&character)
	if err != nil {
		logger.Logger.Error(err.Error())
		return nil, nil, errors.New("failed to load character")
	}

	if lastLocation.City == "" {
		lastLocation.City = player.Hometown
	}

	city := g.World[lastLocation.City]

	if lastLocation.North < 0 || lastLocation.North > int(city.Height) || lastLocation.East < 0 || lastLocation.East > int(city.Width) {
		newLoc := city.RandomLocation()
		lastLocation = &newLoc
	}

	player.Rank = GetRank(player.Reputation)
	player.LastLocation = *lastLocation

	if character.GangId > 0 {
		gang, err := g.GetGang(character.GangId)
		if err != nil {
			logger.Logger.Error(err.Error())
		} else {
			gang.Member = player
			player.Gang = gang
		}
	}

	return player, lastLocation, nil
}

func (g *Game) GetGang(gangId uint64) (*Gang, error) {
	row := g.DbConn.QueryRow(`
        SELECT
            id,
            name,
            tag,
            leader_id,
        FROM
            gangs 
        WHERE
            id = ?
        LIMIT
            1
        `, gangId)

	if row == nil {
		return nil, errors.New("no gang found")
	}

	gangData := models.Gang{}
	err := row.Scan(
		&gangData.Id,
		&gangData.Name,
		&gangData.Tag,
		&gangData.LeaderID,
	)

	if err != nil || gangData.Id == 0 {
		logger.Logger.Error(err.Error())
		return nil, errors.New("no gang found")
	}

	gang := Gang{
		ID:       gangData.Id,
		Name:     gangData.Name,
		Tag:      gangData.Tag,
		LeaderID: gangData.LeaderID,
	}

	return &gang, nil
}
