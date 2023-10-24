package models

type User struct {
	Id        uint64
	Email     string
	Password  string
	UserType  uint64
	CreatedAt uint64
}

type Character struct {
	Id            uint64
	UserId        uint64
	Name          string
	Reputation    int64
	Health        uint
	NpcKills      uint
	PlayerKills   uint
	Cash          uint32
	Bank          uint32
	Hometown      string
	SkillAcc      float32
	SkillHide     float32
	SkillSearch   float32
	SkillTrack    float32
	SkillSnoop    float32
	GangId        uint64
	IsAdmin       int
	LocationNorth int
	LocationEast  int
	LocationCity  string
	CreatedAt     int64
}

type Inventory struct {
	UserId    uint64
	Inventory string
}

type Gang struct {
	Id       uint64
	Name     string
	Tag      string
	LeaderID uint64
}
