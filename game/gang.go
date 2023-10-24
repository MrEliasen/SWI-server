package game

type Gang struct {
	ID       uint64
	Name     string
	Tag      string
	LeaderID uint64
	Member   *Entity
}

func (g *Gang) IsLeader(e *Entity) bool {
	return g.LeaderID == e.PlayerID
}

func (g *Gang) Save() bool {
	if g.ID <= 0 {
		_, err := g.Member.Client.Game.DbConn.Exec(
			"INSERT INTO gangs (name, tag, leader_id ) VALUES (?, ?, ?)",
			g.Name,
			g.Tag,
			g.LeaderID,
		)
		if err != nil {
			print(err.Error())
		}

		return err == nil
	}

	_, err := g.Member.Client.Game.DbConn.Exec(
		"UPDATE gangs SET name = ?, tag = ?, leader_id = ? WHERE id = ?",
		g.Name,
		g.Tag,
		g.LeaderID,
		g.ID,
	)
	if err != nil {
		print(err.Error())
	}

	return err == nil
}
