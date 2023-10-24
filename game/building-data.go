package game

import (
	"github.com/google/uuid"
	"github.com/mreliasen/swi-server/internal/responses"
)

type BuildingTemplate struct {
	Name         string
	Commands     []string
	MerchantType responses.MerchantType
	ShopStock    map[string]int32
	ShopBuyType  map[IType]int32
}

var BuildingTemplates = map[BuildingType]BuildingTemplate{
	BuildingTypeAirport: {
		Name: "International Airport",
		Commands: []string{
			"/travel",
		},
	},
	BuildingTypeHospital: {
		Name: "Private Hospital",
		Commands: []string{
			"/heal",
		},
	},
	BuildingTypeBank: {
		Name: "City Bank",
		Commands: []string{
			"/withdraw",
			"/deposit",
			"/transfer",
		},
	},
	BuildingTypeBar: {
		Name: "Old Speakeasy",
		Commands: []string{
			"/drink",
		},
	},
	BuildingTypePawnShop: {
		Name:         "Pawn Shop",
		MerchantType: responses.MerchantType_MERCHANT_PAWN_SHOP,
		Commands: []string{
			"/shop",
		},
		ShopStock: map[string]int32{
			"smartphone": -1,
		},
		ShopBuyType: map[IType]int32{
			ItemTypeMelee:      -1,
			ItemTypeSmartPhone: -1,
			ItemTypeTrash:      -1,
		},
	},
	BuildingTypeArms: {
		Name:         "Arms Dealer",
		MerchantType: responses.MerchantType_MERCHANT_ARMS_DEALER,
		Commands: []string{
			"/shop",
		},
		ShopStock: map[string]int32{
			"iia_armor":    -1,
			"ii_armor":     -1,
			"iiia_armor":   -1,
			"iii_armor":    -1,
			"iv_armor":     -1,
			"stabvest":     -1,
			"chainmail":    -1,
			"hardarmor":    -1,
			"subsonic":     -1,
			"sdammo":       -1,
			"plusp":        -1,
			"pluspplus":    -1,
			"apammo":       -1,
			"beretta92":    -1,
			"glock22":      -1,
			"sigp320":      -1,
			"sw610":        -1,
			"1911":         -1,
			"ragingbull":   -1,
			"ar-15":        -1,
			"ak47":         -1,
			"scarh":        -1,
			"m82":          -1,
			"brassknuckle": -1,
			"pipewrench":   -1,
			"crowbar":      -1,
			"switchblade":  -1,
			"bbbat":        -1,
			"fireaxe":      -1,
			"machete":      -1,
			"katana":       -1,
			"chainsaw":     -1,
		},
		ShopBuyType: map[IType]int32{
			ItemTypeGun:        -1,
			ItemTypeAmmo:       -1,
			ItemTypeArmor:      -1,
			ItemTypeMelee:      -1,
			ItemTypeSmartPhone: -1,
		},
	},
}

func NewBuilding(t BuildingType) Building {
	template := BuildingTemplates[t]
	buildingCommands := map[string]*Command{}

	for _, cmd := range template.Commands {
		found, ok := BuildingCommandsList[cmd]
		if !ok {
			continue
		}

		buildingCommands[cmd] = found
	}

	stock := []*StockItem{}
	if template.ShopStock != nil {
		for id, amount := range template.ShopStock {
			stock = append(stock, &StockItem{
				TemplateId: id,
				Amount:     amount,
			})
		}
	}

	buyType := make(map[IType]int32)
	if template.ShopBuyType != nil {
		for k, v := range template.ShopBuyType {
			buyType[k] = v
		}
	}

	return Building{
		ID:           uuid.New().String(),
		Name:         template.Name,
		MerchantType: template.MerchantType,
		Commands:     buildingCommands,
		BuildingType: t,
		ShopStock:    stock,
		ShopBuyType:  buyType,
	}
}
