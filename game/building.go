package game

import (
	"sync"

	"github.com/mreliasen/swi-server/internal/responses"
)

type BuildingType = uint8

const (
	BuildingTypeAirport  BuildingType = 1
	BuildingTypeHospital BuildingType = 2
	BuildingTypeBank     BuildingType = 3
	BuildingTypeBar      BuildingType = 4
	BuildingTypeArms     BuildingType = 5
	BuildingTypePawnShop BuildingType = 6
)

type Building struct {
	ID           string
	Name         string
	Description  string
	BuildingType BuildingType
	MerchantType responses.MerchantType
	Commands     map[string]*Command
	ShopStock    []*StockItem
	ShopBuyType  map[IType]int32
	Mu           sync.Mutex
}

type BuildingGameFrame struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Commands    []string `json:"commands"`
}

func (b *Building) gameFrame() *responses.Building {
	cmdList := []string{}

	for cmd := range b.Commands {
		cmdList = append(cmdList, cmd)
	}

	return &responses.Building{
		Name:     b.Name,
		Commands: cmdList,
	}
}
