package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mreliasen/swi-server/internal/responses"
)

func (n *Entity) Restock() {
	totalDrugs := len(DrugsList)

	for i := 0; i < len(n.Inventory.Items); i++ {
		// dont remove equipment
		if item := n.Inventory.Items[i]; item != nil {
			if item.GetItemType() != ItemTypeDrug {
				continue
			}
		}

		itemIndex := rand.Intn(totalDrugs)
		condition := (float32(rand.Intn(100)) + 1) / 100

		if item, ok := NewItem(DrugsList[itemIndex].TemplateName); ok {
			item.Condition = condition
			n.Inventory.addItem(item)
			/* item.Inventory = n.Inventory
			   n.Inventory.Items[i] = item */
		}
	}
}

func (e *Entity) OpenDrugDealer(c *Client) {
	if e.IsPlayer || e.NpcType != DrugDealer {
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

	e.SyncDealerInventory(c)
}

func (e *Entity) SyncDealerInventory(c *Client) {
	if e.IsPlayer || e.NpcType != DrugDealer {
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

	items := []*responses.MerchantItemGroup{
		{
			Items: []*responses.MerchantItem{},
		},
	}

	for index, item := range e.Inventory.Items {
		if item == nil {
			items[0].Items = append(items[0].Items, &responses.MerchantItem{})
			continue
		}

		if item.GetItemType() == ItemTypeDrug {
			price := uint32(float32(item.GetPrice()) * e.Loc.City.DrugDemands[item.TemplateName])

			if price <= 0 {
				price = 1
			}

			items[0].Items = append(items[0].Items, &responses.MerchantItem{
				Name:        item.GetName(),
				Price:       price,
				Description: item.GetDescription() + item.GetQualitySuffix(),
				Condition:   item.Condition,
				Quantity:    item.Amount,
				Demand:      e.Loc.City.DrugDemands[item.TemplateName],
				Index:       int32(index),
			})
		}
	}

	e.Mu.Lock()
	e.ShoppingWith[c.Player] = time.Now().Unix()
	e.Mu.Unlock()
	c.Player.ShoppingWith[e] = time.Now().Unix()

	c.SendEvent(&responses.MerchantInventory{
		MerchantName: e.Name,
		MerchantId:   e.NpcID,
		MerchantType: responses.MerchantType_MERCHANT_DRUG_DEALER,
		Items:        items,
		Open:         true,
	})
}
