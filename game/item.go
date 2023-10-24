package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/mreliasen/swi-server/internal/responses"
	"golang.org/x/exp/slices"
)

type IType uint8

const (
	ItemTypeTrash      = 0
	ItemTypeGun        = 1
	ItemTypeMelee      = 2
	ItemTypeArmor      = 3
	ItemTypeAmmo       = 4
	ItemTypeSmartPhone = 5
	ItemTypeDrug       = 6
	ItemTypeLabel      = 6
	ItemTypeMystery    = 7
)

type Item struct {
	ID           string
	Name         string
	Description  string
	TemplateName string
	ItemType     IType
	Condition    float32
	Amount       int32
	Price        uint32
	MinRep       int64
	BasePrice    uint
	MaxPrice     uint
	Damage       uint
	ArmorGuns    uint
	ArmorMelee   uint
	AmmoWear     float32
	Inventory    *Inventory
	Loc          *Location
	UseEffect    *ItemUseEffect
}

type ItemMoved struct {
	Item   *Item
	By     string
	Player *Entity
}

type ItemGameFrame struct {
	ID           string  `json:"id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Description  string  `json:"description,omitempty"`
	Condition    float32 `json:"condition,omitempty"`
	ItemType     IType   `json:"type,omitempty"`
	Amount       int32   `json:"amount,omitempty"`
	Price        uint32  `json:"price,omitempty"`
	Damage       uint    `json:"damage,omitempty"`
	ArmorGuns    uint    `json:"armor_guns,omitempty"`
	ArmorMelee   uint    `json:"armor_melee,omitempty"`
	Equipped     bool    `json:"equipped,omitempty"`
	IsGear       bool    `json:"is_gear,omitempty"`
	HasUseEffect bool    `json:"has_use_effect,omitempty"`
}

func (g *Item) GameFrame() *responses.Item {
	return &responses.Item{
		Id:           g.ID,
		Amount:       g.Amount,
		Name:         g.GetName(),
		Description:  g.GetDescription(),
		Price:        g.GetPrice(),
		Damage:       uint32(g.GetDamage()),
		ArmorGuns:    uint32(g.GetArmorGuns()),
		ArmorMelee:   uint32(g.GetArmorMelee()),
		HasUseEffect: g.GetUseEffect() != nil,
	}
}

func (g *Item) InspectName() string {
	prefix := ""

	switch g.GetItemType() {
	case ItemTypeDrug:
		prefix = "some "
	default:
		char := string(g.GetName()[0])
		haystack := []string{"a", "i", "e", "o", "u", "y"}

		if slices.Contains(haystack, strings.ToLower(char)) {
			prefix = "an "
		} else {
			prefix = "a "
		}
	}

	return prefix + g.GetName()
}

func (i *Item) GetName() string {
	return ItemsList[i.TemplateName].Name
}

func (i *Item) GetMinRep() int64 {
	return ItemsList[i.TemplateName].MinRep
}

func (i *Item) GetDescription() string {
	return ItemsList[i.TemplateName].Description
}

func (i *Item) GetItemStats() []string {
	desc := []string{}

	if i.GetDamage() > 0 {
		if i.GetItemType() == ItemTypeAmmo {
			desc = append(desc, fmt.Sprintf("Additional Damage: %d", i.GetDamage()))
		} else {
			desc = append(desc, fmt.Sprintf("Damage: %d", i.GetDamage()))
		}
	}

	if i.GetArmorGuns() > 0 {
		desc = append(desc, fmt.Sprintf("Damage Reduction (Firearms): %d", i.GetArmorGuns()))
	}

	if i.GetArmorMelee() > 0 {
		desc = append(desc, fmt.Sprintf("Damage REduction (Melee): %d", i.GetArmorGuns()))
	}

	if i.GetItemType() == ItemTypeAmmo || i.Amount > 1 {
		if i.GetItemType() == ItemTypeAmmo {
			desc = append(desc, fmt.Sprintf("Rounds Left: %d", i.Amount))
		} else {
			desc = append(desc, fmt.Sprintf("Amount/Uses: %d", i.Amount))
		}
	}

	if i.GetItemType() == ItemTypeTrash {
		desc = append(desc, ": The Pawn Shop might be interested in this..")
	}

	return desc
}

func (i *Item) GetQualitySuffix() string {
	if i.GetItemType() == ItemTypeDrug {
		switch {
		case i.Condition > 0.90:
			return " of excellent quality"

		case i.Condition > 0.80:
			return " of good quality"

		case i.Condition > 0.70:
			return " of decent quality"

		case i.Condition > 0.60:
			return " of mixed quality"

		case i.Condition > 0.50:
			return " of useable quality"

		case i.Condition > 0.40:
			return " of poor quality"

		case i.Condition > 0.30:
			return " of bad quality"

		case i.Condition > 0:
			return " of terrible quality"

		default:
			return " of unknown quality"
		}
	}

	switch {
	case i.Condition > 0.90:
		return " in excellent condition"

	case i.Condition > 0.80:
		return " in good condition"

	case i.Condition > 0.70:
		return " in decent condition"

	case i.Condition > 0.60:
		return " in reasonable condition"

	case i.Condition > 0.50:
		return " in questionable condition"

	case i.Condition > 0.40:
		return " in poor condition"

	case i.Condition > 0.30:
		return " in bad condition"

	case i.Condition > 0:
		return " in terrible condition"

	default:
		return ""
	}
}

func (i *Item) GetItemType() IType {
	return ItemsList[i.TemplateName].ItemType
}

func (i *Item) GetPrice() uint32 {
	if i.GetItemType() == ItemTypeDrug {
		min := float32(ItemsList[i.TemplateName].BasePrice)
		max := float32(ItemsList[i.TemplateName].MaxPrice)

		if i.GetItemType() == ItemTypeDrug {
			return uint32(min * i.Condition)
		}

		return uint32((max - min*i.Condition) + min)
	}

	return uint32(math.Floor((float64(ItemsList[i.TemplateName].BasePrice) / float64(ItemsList[i.TemplateName].Amount)) * float64(i.Amount)))
}

func (i *Item) GetDamage() uint {
	return ItemsList[i.TemplateName].Damage
}

func (i *Item) GetArmorGuns() uint {
	return ItemsList[i.TemplateName].ArmorGuns
}

func (i *Item) GetArmorMelee() uint {
	return ItemsList[i.TemplateName].ArmorMelee
}

func (i *Item) GetUseEffect() *ItemUseEffect {
	return ItemsList[i.TemplateName].UseEffect
}

func (i *Item) GetAmmoWear() float32 {
	return ItemsList[i.TemplateName].AmmoWear
}

func (i *Item) IsGear() bool {
	if i.GetItemType() == ItemTypeGun {
		return true
	}
	if i.GetItemType() == ItemTypeMelee {
		return true
	}
	if i.GetItemType() == ItemTypeAmmo {
		return true
	}
	if i.GetItemType() == ItemTypeArmor {
		return true
	}

	return false
}
