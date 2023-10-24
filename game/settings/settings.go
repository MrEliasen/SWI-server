package settings

import "time"

const (
	// websocket settings
	WriteWait      = 10 * time.Second
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 256

	// Misc settings
	// ClientVersionCheck = 202310082253
	// MaxPlayers         = 100
	AutoSaveMinutes   = 5
	CombatLoggingSecs = 10

	// commands / actions
	HealCostPerPoint  = 30
	DrinkRepGain      = 5
	DrinkCost         = 100
	DrinkHealthCost   = 5
	DrinkSkillCost    = 0.001
	SmartPhoneCost    = 25
	DrugUseRepGain    = 1
	DrugUseHealthCost = 9

	// NPC and game Timers
	CityDemandUpdateMinMins   = 45
	CityDemandUpdateMaxMins   = 90
	DrugRestockDelaySeconds   = 60 * 20
	NpcRespawnDelaySeconds    = 60 * 15
	DealerRespawnDelaySeconds = 60 * 15
	NPCMoveMaxDelaySeconds    = 120
	NPCMoveMinDelaySeconds    = 30
	NPCAttackDelayMs          = 2250
	DroppedItemsDecaySeconds  = 60 * 60
	TravelCostChangeMinutes   = 60

	// skills
	PlayerAttackDelayMs = 2200
	PlayerSkillDelayMs  = 2200
	PlayerMoveDelayMs   = 150

	// General Player Stats and timers
	PlayerHealDelaySeconds = 60
	PlayerMaxHealth        = 100
	PlayerMaxInventory     = 30
	DrugProfitMargin       = 1.2
	DrugRepIncrease        = 3
	ItemSellPriceLoss      = 0.65

	// New players
	PlayerStartCash        = 100
	PlayerStartBank        = 500
	PlayerStartReputation  = 0
	PlayerStartSkillAcc    = 20
	PlayerStartSkillTrack  = 1
	PlayerStartSkillSnoop  = 1
	PlayerStartSkillHide   = 1
	PlayerStartSkillSearch = 1
)
