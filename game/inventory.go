package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal"
	"github.com/mreliasen/swi-server/internal/database/models"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/mreliasen/swi-server/internal/responses"
)

type Inventory struct {
	Items     [settings.PlayerMaxInventory]*Item `json:"items"`
	Owner     *Entity
	Equipment map[IType]*Item
	Mu        sync.Mutex
}

type ItemSaveContainer struct {
	ID           string  `json:"id"`
	TemplateName string  `json:"template_name"`
	Condition    float32 `json:"condition"`
	Amount       uint    `json:"amount"`
	Equipped     bool    `json:"equipped"`
}

func NewInventory(owner *Entity) *Inventory {
	return &Inventory{
		Items: [settings.PlayerMaxInventory]*Item{},
		Owner: owner,
		Equipment: map[IType]*Item{
			ItemTypeGun:   nil,
			ItemTypeAmmo:  nil,
			ItemTypeArmor: nil,
		},
	}
}

func (i *Inventory) Info(itemId string) {
	for _, item := range i.Items {
		if item == nil {
			continue
		}

		if item.ID == itemId {
			conditionSuffix := ""

			if item.GetItemType() == ItemTypeGun {
				logger.Logger.Trace(fmt.Sprintf("%f", item.Condition))
				conditionSuffix = ". " + fmt.Sprintf("Risk of Jamming: %.2f", (1.0-item.Condition)*100) + "%"
			}

			output := [][]string{
				{"Condition", strings.TrimSpace(item.GetQualitySuffix()) + conditionSuffix},
			}

			stats := item.GetItemStats()

			if len(stats) > 0 {
				for _, stat := range stats {
					parts := strings.Split(stat, ": ")
					name := parts[0]
					value := strings.Join(parts[1:], ": ")
					output = append(output, []string{name, value})
				}
			}

			i.Owner.Client.SendEvent(&responses.Generic{
				Ascii:    true,
				Messages: internal.ToTable([]string{"Item", item.GetName()}, output),
			})
			return
		}
	}
}

func (i *Inventory) equip(itemId string) bool {
	i.Mu.Lock()
	defer i.Mu.Unlock()

	for _, item := range i.Items {
		if item == nil {
			continue
		}

		if item.ID == itemId {
			itype := item.GetItemType()

			if itype != ItemTypeArmor && itype != ItemTypeGun && itype != ItemTypeAmmo && itype != ItemTypeMelee {
				logger.Logger.Warn(fmt.Sprintf("Item %v is not an equippble type %v.", item.GetName(), itype))
				return false
			}

			i.Equipment[item.GetItemType()] = item
			return true
		}
	}

	return false
}

func (i *Inventory) unequip(slot int) {
	i.Mu.Lock()
	defer i.Mu.Unlock()

	item := i.Items[slot]
	if item == nil {
		return
	}

	itemType := item.GetItemType()

	if itemType != ItemTypeArmor && itemType != ItemTypeGun && itemType != ItemTypeAmmo {
		return
	}

	i.Equipment[itemType] = nil
}

func (inv *Inventory) addItem(x *Item) error {
	inv.Mu.Lock()
	defer inv.Mu.Unlock()

	for i, itemSlot := range inv.Items {
		if itemSlot == nil {
			x.Inventory = inv
			x.Loc = nil
			inv.Items[i] = x

			if inv.Owner.IsPlayer && x.TemplateName == "smartphone" {
				go inv.Owner.PlayerSendMapUpdate()
			}

			return nil
		}
	}

	return errors.New("no more inventory space left")
}

func (inv *Inventory) HasItem(templateId string) (bool, *Item) {
	for _, item := range inv.Items {
		if item != nil && item.TemplateName == templateId {
			return true, item
		}
	}

	return false, nil
}

func (inv *Inventory) GameFrame() *responses.Inventory {
	event := responses.Inventory{
		Items: []*responses.Item{},
	}

	for _, item := range inv.Items {
		if item != nil {
			frame := item.GameFrame()
			frame.Equipped = inv.Equipment[item.GetItemType()] == item
			frame.IsGear = item.IsGear()
			event.Items = append(event.Items, frame)
		} else {
			event.Items = append(event.Items, &responses.Item{})
		}
	}

	return &event
}

func (inv *Inventory) dump(l *Location) {
	if len(inv.Items) == 0 {
		return
	}

	for slot := range inv.Equipment {
		inv.Equipment[slot] = nil
	}

	for index := range inv.Items {
		event := inv.drop(index)

		if event != nil && event.Item != nil {
			logger.LogItems(inv.Owner.Name, "dump", event.Item.TemplateName, inv.Owner.Loc.Coords.North, inv.Owner.LastLocation.East, inv.Owner.Loc.City.ShortName)

			l.AddItem <- event
		}
	}
}

func (inv *Inventory) HasRoom() bool {
	if len(inv.Items) == 0 {
		return false
	}

	free := false
	for _, ref := range inv.Items {
		if ref == nil {
			free = true
			break
		}
	}

	return free
}

func (inv *Inventory) drop(i int) *ItemMoved {
	inv.Mu.Lock()
	defer inv.Mu.Unlock()

	if inv.Items[i] == nil {
		return nil
	}

	// remove from inventory
	item := inv.Items[i]
	inv.Items[i] = nil

	// remove inv ref from item
	item.Inventory = nil

	// place item on the ground at current location
	var loc *Location
	var DroppedBy string
	var player *Entity

	if inv.Owner.IsPlayer {
		loc = inv.Owner.Loc
		player = inv.Owner
		DroppedBy = player.Name
	} else {
		loc = inv.Owner.Loc
		DroppedBy = inv.Owner.Name
	}

	if inv.Equipment[item.GetItemType()] == item {
		inv.Equipment[item.GetItemType()] = nil
	}

	if loc == nil {
		return nil
	}

	if player != nil {
		player.PlayerSendInventoryUpdate()

		if item.TemplateName == "smartphone" {
			go player.PlayerSendMapUpdate()
		}
	}

	return &ItemMoved{
		Item:   item,
		By:     DroppedBy,
		Player: player,
	}
}

func (i *Inventory) save() bool {
	if i.Owner == nil {
		return false
	}

	i.Mu.Lock()
	defer i.Mu.Unlock()

	payload := []ItemSaveContainer{}

	for _, item := range i.Items {
		if item == nil {
			payload = append(payload, ItemSaveContainer{})
			continue
		}

		payload = append(payload, ItemSaveContainer{
			ID:           item.ID,
			Amount:       uint(item.Amount),
			TemplateName: item.TemplateName,
			Condition:    item.Condition,
			Equipped:     i.Equipment[item.GetItemType()] == item,
		})
	}

	val, err := json.Marshal(payload)
	if err != nil {
		print(err.Error())
		return false
	}

	_, err = i.Owner.Client.Game.DbConn.Exec(
		"INSERT OR REPLACE INTO inventory (user_id, inventory, updated_at) VALUES(?, ?, ?)",
		i.Owner.Client.UserId, string(val), time.Now().Unix(),
	)
	if err != nil {
		print(err.Error())
	}

	return err == nil
}

func (i *Inventory) load() {
	i.Mu.Lock()
	defer i.Mu.Unlock()

	if !i.Owner.IsPlayer {
		return
	}

	row := i.Owner.Client.Game.DbConn.QueryRow("SELECT inventory FROM inventory WHERE user_id = ? LIMIT 1", i.Owner.Client.UserId)
	invData := models.Inventory{}
	row.Scan(&invData.Inventory)

	if invData.Inventory == "" {
		return
	}

	items := []ItemSaveContainer{}
	json.Unmarshal([]byte(invData.Inventory), &items)

	for index, item := range items {
		if item.ID == "" {
			continue
		}

		if newItem, ok := NewItem(item.TemplateName); ok {
			newItem.ID = item.ID
			newItem.Condition = item.Condition
			newItem.Amount = int32(item.Amount)
			newItem.Inventory = i
			i.Items[index] = newItem

			if item.Equipped {
				i.Equipment[newItem.GetItemType()] = newItem
			}
		}
	}
}

func (inv *Inventory) useById(id string) {
	for i, item := range inv.Items {
		if item != nil && item.ID == id {
			inv.use(i)
			return
		}
	}
}

func (inv *Inventory) use(i int) {
	if inv.Items[i] == nil || inv.Items[i].GetUseEffect() == nil {
		return
	}

	logger.LogItems(inv.Owner.Name, "use", inv.Items[i].TemplateName, inv.Owner.Loc.Coords.North, inv.Owner.LastLocation.East, inv.Owner.Loc.City.ShortName)

	inv.Items[i].GetUseEffect().Use(inv.Owner.Client, inv.Items[i], i)
}
