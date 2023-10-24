package game

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mreliasen/swi-server/game/skills"
	"github.com/mreliasen/swi-server/internal/logger"
)

const (
	DrugDealer       NPCType = 0
	DrugAddict       NPCType = 1
	Homeless         NPCType = 2
	Tweaker          NPCType = 3
	Bouncer          NPCType = 4
	Busker           NPCType = 5
	StreetVendor     NPCType = 6
	StreetGangMember NPCType = 7
	Tourist          NPCType = 8
	Activist         NPCType = 9
	PoliceOfficer    NPCType = 10
	BeatCop          NPCType = 11
	DeliveryDriver   NPCType = 12
)

type NpcTemplate struct {
	NpcType   NPCType
	Title     string
	Rep       int32
	Cash      uint32
	Gender    Gender
	Health    int
	SkillAcc  float32
	Equipment map[IType]string
	Inventory []string
	commands  map[string]*Command
}

func NewNPC(npcType NPCType) *Entity {
	var name string

	id, err := uuid.NewUUID()
	npcId := fmt.Sprintf("%d", time.Now().UnixNano())

	if err == nil {
		npcId = id.String()
	}

	npc := Entity{
		NpcID:         npcId,
		Name:          name,
		NpcType:       npcType,
		NpcRepReward:  NpcTemplates[npcType].Rep,
		NpcCashReward: NpcTemplates[npcType].Cash,
		Health:        NpcTemplates[npcType].Health,
		NpcGender:     NpcTemplates[npcType].Gender,
		IsPlayer:      false,
		NpcTitle:      NpcTemplates[npcType].Title,
		NpcCommands:   NpcTemplates[npcType].commands,
		SkillAcc: skills.Accuracy{
			Value: NpcTemplates[npcType].SkillAcc,
		},
		ShoppingWith: make(map[*Entity]int64),
		NpcHostiles:  make(map[string]bool),
		TargetedBy:   make(map[*Entity]bool),
	}

	npc.RandomiseGenderName()
	npc.Inventory = NewInventory(&npc)

	if NpcTemplates[npcType].Equipment != nil {
		for _, tmplId := range NpcTemplates[npcType].Equipment {
			newItem, ok := NewItem(tmplId)
			if !ok {
				continue
			}

			npc.Inventory.addItem(newItem)
			ok = npc.Inventory.equip(newItem.ID)

			if !ok {
				logger.Logger.Fatal(fmt.Sprintf("Failed to equip slot: %s for NPC: %s", tmplId, npc.NpcTitle))
			}
		}
	}

	for _, tmplId := range NpcTemplates[npcType].Inventory {
		newItem, ok := NewItem(tmplId)
		if !ok {
			continue
		}

		npc.Inventory.addItem(newItem)
	}

	npc.NPCStartRoutines()
	return &npc
}

var NpcTemplates = map[NPCType]NpcTemplate{
	DrugDealer: {
		NpcType:  DrugDealer,
		Title:    "Drug Dealer",
		Health:   500,
		Cash:     1000,
		Rep:      -1750,
		SkillAcc: 75.0,
		Gender:   GenderMale,
		Equipment: map[IType]string{
			ItemTypeArmor: "iiia_armor",
			ItemTypeGun:   "deserteagle",
			ItemTypeAmmo:  "plusp",
		},
	},
	DrugAddict: {
		NpcType:  DrugAddict,
		Title:    "Drug Addict",
		Gender:   GenderRandom,
		Health:   300,
		Cash:     500,
		Rep:      -1750,
		SkillAcc: 75.0,
		Equipment: map[IType]string{
			ItemTypeArmor: "iiia_armor",
			ItemTypeGun:   "ragingbull",
			ItemTypeAmmo:  "plusp",
		},
	},
	Homeless: {
		NpcType:  Homeless,
		Title:    "Homeless",
		Gender:   GenderRandom,
		Health:   80,
		Cash:     12,
		Rep:      25,
		SkillAcc: 15.0,
		Inventory: []string{
			"crack",
			"ketamine",
		},
		Equipment: map[IType]string{
			ItemTypeMelee: "brokenbottle",
		},
	},
	Tweaker: {
		NpcType:  Tweaker,
		Title:    "Tweaker",
		Gender:   GenderRandom,
		Health:   100,
		Cash:     25,
		Rep:      140,
		SkillAcc: 35.0,
		Inventory: []string{
			"meth",
			"meth",
		},
		Equipment: map[IType]string{
			ItemTypeMelee: "crowbar",
		},
	},
	Bouncer: {
		NpcType:  Bouncer,
		Title:    "Bouncer",
		Gender:   GenderRandom,
		Health:   110,
		Cash:     200,
		SkillAcc: 45.0,
		Rep:      272,
		Inventory: []string{
			"coke",
			"goldchain",
		},
		Equipment: map[IType]string{
			ItemTypeMelee: "bbat",
			ItemTypeArmor: "stabvest",
		},
	},
	Busker: {
		NpcType:  Busker,
		Title:    "Busker",
		Gender:   GenderRandom,
		Health:   100,
		Cash:     70,
		SkillAcc: 35,
		Rep:      70,
		Inventory: []string{
			"festivalticket",
			"weed",
		},
		Equipment: map[IType]string{
			ItemTypeMelee: "guitar",
		},
	},
	StreetVendor: {
		NpcType:  StreetVendor,
		Title:    "Street Vendor",
		Gender:   GenderRandom,
		Health:   140,
		Cash:     80,
		SkillAcc: 40,
		Rep:      70,
		Inventory: []string{
			"sunglasses",
		},
		Equipment: map[IType]string{
			ItemTypeMelee: "leadpipe",
		},
	},
	DeliveryDriver: {
		NpcType:  DeliveryDriver,
		Title:    "Delivery Driver",
		Gender:   GenderRandom,
		Health:   100,
		Cash:     10,
		SkillAcc: 40,
		Rep:      80,
		Inventory: []string{
			"deliverypackage",
			"smartphone",
		},
		Equipment: map[IType]string{
			ItemTypeMelee: "parcel",
		},
	},
	Tourist: {
		NpcType:  Tourist,
		Title:    "Tourist",
		Gender:   GenderRandom,
		Health:   65,
		Cash:     10,
		SkillAcc: 25,
		Rep:      33,
		Inventory: []string{
			"smartphone",
			"sunglasses",
		},
		Equipment: map[IType]string{},
	},
	StreetGangMember: {
		NpcType:  StreetGangMember,
		Title:    "Street Gang Member",
		Gender:   GenderRandom,
		Health:   150,
		Cash:     400,
		SkillAcc: 50,
		Rep:      412,
		Inventory: []string{
			"sdammo",
			"sdammo",
			"sdammo",
			"coke",
			"coke",
			"weed",
		},
		Equipment: map[IType]string{
			ItemTypeGun:   "mac-10",
			ItemTypeAmmo:  "sdammo",
			ItemTypeArmor: "ii_armor",
		},
	},
	Activist: {
		NpcType:  Activist,
		Title:    "Activist",
		Gender:   GenderRandom,
		Health:   40,
		Cash:     20,
		SkillAcc: 10,
		Rep:      20,
		Inventory: []string{
			"currentthing",
			"weed",
		},
		Equipment: map[IType]string{
			ItemTypeMelee: "bikelock",
		},
	},
	BeatCop: {
		NpcType:  BeatCop,
		Title:    "Beat Cop",
		Gender:   GenderRandom,
		Health:   100,
		Cash:     30,
		SkillAcc: 50,
		Rep:      200,
		Inventory: []string{
			"policebadge",
			"sdammo",
			"sdammo",
		},
		Equipment: map[IType]string{
			ItemTypeGun:   "glock22",
			ItemTypeAmmo:  "sdammo",
			ItemTypeArmor: "iia_armor",
		},
	},
	PoliceOfficer: {
		NpcType:  PoliceOfficer,
		Title:    "Police Officer",
		Gender:   GenderRandom,
		Health:   130,
		Cash:     30,
		SkillAcc: 60,
		Rep:      546,
		Inventory: []string{
			"policebadge",
			"sdammo",
			"sdammo",
		},
		Equipment: map[IType]string{
			ItemTypeGun:   "1911",
			ItemTypeAmmo:  "sdammo",
			ItemTypeArmor: "ii_armor",
		},
	},
}
