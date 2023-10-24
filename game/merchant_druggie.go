package game

import (
	"fmt"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal/responses"
)

func (n *Entity) ClearStock() {
	for i := 0; i < len(n.Inventory.Items); i++ {
		// dont remove equipment
		if item := n.Inventory.Items[i]; item != nil {
			if item.GetItemType() != ItemTypeDrug {
				continue
			}
		}

		n.Inventory.drop(i)
	}
}

func (e *Entity) OpenDruggieMenu(c *Client) {
	if e.IsPlayer || e.NpcType != DrugAddict {
		return
	}

	if isHostile := e.NpcHostiles[c.UUID]; isHostile {
		return
	}

	if e.CurrentTarget != nil {
		c.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
			Messages: []string{fmt.Sprintf("%s is in a fight and ignores you.", e.Name)},
		})
		return
	}

	e.SyncDruggieTrade(c)
}

func (e *Entity) SyncDruggieTrade(c *Client) {
	if e.IsPlayer || e.NpcType != DrugAddict {
		return
	}

	if isHostile := e.NpcHostiles[c.UUID]; isHostile {
		return
	}

	if e.CurrentTarget != nil {
		c.SendEvent(&responses.Generic{
			Status:   responses.ResponseStatus_RESPONSE_STATUS_WARN,
			Messages: []string{fmt.Sprintf("%s is in a fight and ignores you.", e.Name)},
		})
		return
	}

	items := []*responses.MerchantItem{}

	for _, item := range c.Player.Inventory.Items {
		if item == nil {
			items = append(items, &responses.MerchantItem{})
			continue
		}

		if item.GetItemType() != ItemTypeDrug {
			items = append(items, &responses.MerchantItem{
				Name:    item.GetName(),
				Cansell: false,
			})
			continue
		}

		price := uint32((float32(item.GetPrice()) * settings.DrugProfitMargin) * e.Loc.City.DrugDemands[item.TemplateName])

		if price <= 0 {
			price = 1
		}

		items = append(items, &responses.MerchantItem{
			Name:        item.GetName(),
			Price:       price,
			Description: item.GetDescription() + item.GetQualitySuffix(),
			Condition:   item.Condition,
			Quantity:    item.Amount,
			Demand:      e.Loc.City.DrugDemands[item.TemplateName],
			Cansell:     true,
		})
	}

	e.Mu.Lock()
	e.ShoppingWith[c.Player] = time.Now().Unix()
	e.Mu.Unlock()
	c.Player.ShoppingWith[e] = time.Now().Unix()

	c.SendEvent(&responses.MerchantInventory{
		MerchantName:   e.Name,
		MerchantId:     e.NpcID,
		MerchantGender: responses.Gender(e.NpcGender),
		MerchantType:   responses.MerchantType_MERCHANT_DRUGGIE,
		PlayerItems:    items,
		Open:           true,
	})
}
