package game

import (
	"strings"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal/responses"
)

type StockItem struct {
	TemplateId string
	Amount     int32
}

func (e *Building) OpenShop(c *Client) {
	e.SyncShopInventory(c)
}

func (e *Building) SyncShopInventory(c *Client) {
	items := map[uint32]*responses.MerchantItemGroup{}

	for stockIndex, stock := range e.ShopStock {
		item := ItemsList[stock.TemplateId]
		itype32 := uint32(item.ItemType)

		if _, ok := items[itype32]; !ok {
			items[itype32] = &responses.MerchantItemGroup{
				Type:  responses.ItemType(itype32),
				Items: []*responses.MerchantItem{},
			}
		}

		items[itype32].Items = append(items[itype32].Items, &responses.MerchantItem{
			Name:        item.GetName(),
			Price:       item.GetPrice(),
			Description: item.GetDescription() + item.GetQualitySuffix() + "\n" + strings.Join(item.GetItemStats(), "\n"),
			Condition:   1.0,
			Quality:     item.GetQualitySuffix(),
			Canbuy:      c.Player.Reputation >= item.GetMinRep(),
			Index:       int32(stockIndex),
			Rep:         item.GetMinRep(),
			Quantity:    stock.Amount,
		})
	}

	merchantItems := []*responses.MerchantItemGroup{}
	for _, group := range items {
		merchantItems = append(merchantItems, group)
	}

	playerItems := []*responses.MerchantItem{}

	for _, item := range c.Player.Inventory.Items {
		if item == nil {
			playerItems = append(playerItems, &responses.MerchantItem{})
			continue
		}

		if item.GetItemType() == ItemTypeDrug {
			playerItems = append(playerItems, &responses.MerchantItem{
				Name:    item.GetName(),
				Cansell: false,
			})
			continue
		}

		playerItems = append(playerItems, &responses.MerchantItem{
			Name:        item.GetName(),
			Price:       uint32(float32(item.GetPrice()) * settings.ItemSellPriceLoss),
			Description: item.GetDescription() + item.GetQualitySuffix() + "\n" + strings.Join(item.GetItemStats(), "\n"),
			Condition:   item.Condition,
			Quantity:    item.Amount,
			Cansell:     true,
		})
	}

	event := &responses.MerchantInventory{
		MerchantName: e.Name,
		MerchantId:   e.ID,
		MerchantType: responses.MerchantType(e.MerchantType),
		Items:        merchantItems,
		PlayerItems:  playerItems,
		Open:         true,
	}

	c.SendEvent(event)
}
