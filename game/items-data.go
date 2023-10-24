package game

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mreliasen/swi-server/game/settings"
)

type ItemTemplate struct {
	Name        string
	Description string
	ItemType    IType
	Uses        bool
	BasePrice   uint
	MaxPrice    uint
	MinRep      int64
	Damage      uint
	Amount      uint
	AmmoWear    float32
	ArmorGuns   uint
	ArmorMelee  uint
	UseEffect   string
}

func NewItem(itemId string) (*Item, bool) {
	baseItem, ok := ItemsList[itemId]

	if !ok {
		return nil, false
	}

	return &Item{
		ID:           uuid.New().String(),
		Amount:       1,
		TemplateName: baseItem.TemplateName,
		Condition:    1.00,
	}, true
}

var (
	ItemsList = map[string]*Item{}
	DrugsList = []*Item{}
)

func GenerateItemsList() {
	for k, item := range ItemTemplates {
		newItem := &Item{
			Name:         item.Name,
			Description:  item.Description,
			TemplateName: k,
			ItemType:     item.ItemType,
			Amount:       1,
			BasePrice:    item.BasePrice,
			MaxPrice:     item.MaxPrice,
			Damage:       item.Damage,
			MinRep:       item.MinRep,
			AmmoWear:     item.AmmoWear,
			ArmorGuns:    item.ArmorGuns,
			ArmorMelee:   item.ArmorMelee,
		}

		if item.UseEffect != "" {
			effect := UseEffectsList[item.UseEffect]
			newItem.UseEffect = effect
		}

		ItemsList[k] = newItem
		if newItem.ItemType == ItemTypeDrug {
			DrugsList = append(DrugsList, newItem)
		}
	}
}

var ItemTemplates = map[string]ItemTemplate{
	// drugs
	"crack": {
		Name:        "Crack",
		Description: "1 Gram",
		ItemType:    ItemTypeDrug,
		BasePrice:   30,
		MaxPrice:    60,
		UseEffect:   "usedrug",
	},
	"coke": {
		Name:        "Cocaine",
		Description: "1 Gram",
		ItemType:    ItemTypeDrug,
		BasePrice:   60,
		MaxPrice:    100,
		UseEffect:   "usedrug",
	},
	"heroine": {
		Name:        "Heroin",
		Description: "1 Gram",
		ItemType:    ItemTypeDrug,
		BasePrice:   150,
		MaxPrice:    500,
		UseEffect:   "usedrug",
	},
	"meth": {
		Name:        "Meth",
		Description: "1 Gram",
		ItemType:    ItemTypeDrug,
		BasePrice:   20,
		MaxPrice:    40,
		UseEffect:   "usedrug",
	},
	"weed": {
		Name:        "Weed",
		Description: "7 Grams / a quater ounce",
		ItemType:    ItemTypeDrug,
		BasePrice:   25,
		MaxPrice:    50,
		UseEffect:   "usedrug",
	},
	"fentanyl": {
		Name:        "Fentanyl",
		Description: "1 Pill",
		ItemType:    ItemTypeDrug,
		BasePrice:   25,
		MaxPrice:    50,
		UseEffect:   "usedrug",
	},
	"pcp": {
		Name:        "PCP",
		Description: "1 tablet",
		ItemType:    ItemTypeDrug,
		BasePrice:   5,
		MaxPrice:    15,
		UseEffect:   "usedrug",
	},
	"ketamine": {
		Name:        "Ketamine",
		Description: "A dose",
		ItemType:    ItemTypeDrug,
		BasePrice:   20,
		MaxPrice:    30,
		UseEffect:   "usedrug",
	},

	// ARMOR RANGED
	"iia_armor": {
		Name:        "Level IIA Body Armor",
		Description: "Designed to protect against 9mm.",
		ItemType:    ItemTypeArmor,
		BasePrice:   300,
		ArmorGuns:   3,
		MinRep:      RanksList[6].MinRep,
	},
	"ii_armor": {
		Name:        "Level II Body Armor",
		Description: "Provides protection against upto .357 Magnum.",
		ItemType:    ItemTypeArmor,
		BasePrice:   700,
		ArmorGuns:   7,
		MinRep:      RanksList[10].MinRep,
	},
	"iiia_armor": {
		Name:        "Level IIIA Body Armor",
		Description: "Designed to protect against handgun threats, including .44 Magnum.",
		ItemType:    ItemTypeArmor,
		BasePrice:   1100,
		ArmorGuns:   11,
		MinRep:      RanksList[14].MinRep,
	},
	"iii_armor": {
		Name:        "Level III Body Armor",
		Description: "Designed to protect against rifle threats, including 7.62x51mm (M80) ball.",
		ItemType:    ItemTypeArmor,
		BasePrice:   1500,
		ArmorGuns:   15,
		MinRep:      RanksList[18].MinRep,
	},
	"iv_armor": {
		Name:        "Level IV Body Armor",
		Description: "Provides protection against armor-piercing rifle rounds, such as .30-06 M2 AP.",
		ItemType:    ItemTypeArmor,
		BasePrice:   1900,
		ArmorGuns:   19,
		MinRep:      RanksList[22].MinRep,
	},

	// GUNS
	"beretta92": {
		Name:        "Beretta 92",
		Description: "A widely used semi-automatic pistol in 9mm, known for its accuracy and reliability.",
		ItemType:    ItemTypeGun,
		BasePrice:   300,
		Damage:      3,
		MinRep:      RanksList[5].MinRep,
	},
	"glock22": {
		Name:        "Glock 22",
		Description: "A popular law enforcement pistol chambered in .40 S&W, favored for its versatility.",
		ItemType:    ItemTypeGun,
		BasePrice:   500,
		Damage:      5,
		MinRep:      RanksList[7].MinRep,
	},
	"sigp320": {
		Name:        "Sig Sauer P220",
		Description: "The P220 is known for its reliability and is a favorite among enthusiasts. Chambered in 10mm",
		ItemType:    ItemTypeGun,
		BasePrice:   700,
		Damage:      7,
		MinRep:      RanksList[9].MinRep,
	},
	"sw610": {
		Name:        "Smith & Wesson Model 610",
		Description: "A versatile and durable stainless steel semi-automatic revolver chambered for .357 Magnum.",
		ItemType:    ItemTypeGun,
		BasePrice:   900,
		Damage:      9,
		MinRep:      RanksList[11].MinRep,
	},
	"1911": {
		Name:        "Colt 1911",
		Description: "One of the most iconic handguns in the world. Chambered in .45 ACP",
		ItemType:    ItemTypeGun,
		BasePrice:   1100,
		Damage:      11,
		MinRep:      RanksList[13].MinRep,
	},
	"ragingbull": {
		Name:        "Taurus Raging Bull",
		Description: "A large-framed revolver recognized for its power and ruggedness. Chambered in .44 Magnum.",
		ItemType:    ItemTypeGun,
		BasePrice:   1300,
		Damage:      13,
		MinRep:      RanksList[15].MinRep,
	},
	"ar-15": {
		Name:        "AR-15 Rifle",
		Description: "The AR-15 is a widely used semi-automatic rifle known for its modularity and adaptability. Chambered in 5.56x45mm NATO",
		ItemType:    ItemTypeGun,
		BasePrice:   1500,
		Damage:      15,
		MinRep:      RanksList[17].MinRep,
	},
	"ak47": {
		Name:        "AK-47",
		Description: "The AK-47 is a legendary and rugged assault rifle known for its reliability. Chambered in 7.62x39mm",
		ItemType:    ItemTypeGun,
		BasePrice:   1700,
		Damage:      17,
		MinRep:      RanksList[19].MinRep,
	},
	"scarh": {
		Name:        "FN SCAR-H",
		Description: "A versatile battle rifle used by military forces around the world. Chambered in 7.62x51mm NATO",
		ItemType:    ItemTypeGun,
		BasePrice:   1900,
		Damage:      19,
		MinRep:      RanksList[21].MinRep,
	},
	"m82": {
		Name:        "Barrett M82",
		Description: "The Barrett M82 is a renowned anti-materiel rifle chambered in .50 BMG.",
		ItemType:    ItemTypeGun,
		BasePrice:   2100,
		Damage:      21,
		MinRep:      RanksList[23].MinRep,
	},

	// AMMO
	"subsonic": {
		Name:        "Subsonic Ammo",
		Description: "Subsonic rounds are typically lower in power due to their reduced velocity, making them quieter.",
		ItemType:    ItemTypeAmmo,
		BasePrice:   100,
		Damage:      1,
		AmmoWear:    0.0025,
		MinRep:      RanksList[5].MinRep,
		Amount:      15,
	},
	"sdammo": {
		Name:        "Standard Ammo",
		Description: "Standard ammunition, often referred to as \"ball\" ammunition, is the baseline cartridge for a particular caliber.",
		ItemType:    ItemTypeAmmo,
		BasePrice:   300,
		Damage:      3,
		AmmoWear:    0.0040,
		MinRep:      RanksList[9].MinRep,
		Amount:      15,
	},
	"plusp": {
		Name:        "+P Ammo",
		Description: "Cartridges labeled as +P are loaded with higher powder charges than standard loads of the same caliber.",
		ItemType:    ItemTypeAmmo,
		BasePrice:   500,
		Damage:      5,
		AmmoWear:    0.0055,
		MinRep:      RanksList[13].MinRep,
		Amount:      15,
	},
	"pluspplus": {
		Name:        "+P+ Ammo",
		Description: "These cartridges are loaded with significantly more powder, further increasing muzzle velocity and energy.",
		ItemType:    ItemTypeAmmo,
		BasePrice:   700,
		Damage:      7,
		AmmoWear:    0.0070,
		MinRep:      RanksList[17].MinRep,
		Amount:      15,
	},
	"apammo": {
		Name:        "AP Ammo",
		Description: "These cartridges are loaded with significantly more powder, further increasing muzzle velocity and energy.",
		ItemType:    ItemTypeAmmo,
		BasePrice:   900,
		Damage:      9,
		AmmoWear:    0.0085,
		MinRep:      RanksList[21].MinRep,
		Amount:      15,
	},

	// melee
	"brassknuckle": {
		Name:        "Brass Knuckle",
		Description: "Brass knuckles, a classic street weapon, provide a discreet and formidable edge in close combat.",
		ItemType:    ItemTypeMelee,
		BasePrice:   300,
		Damage:      3,
		MinRep:      RanksList[6].MinRep,
	},
	"pipewrench": {
		Name:        "Pipe Wrench",
		Description: "A heavy-duty pipe wrench, perfect for tight spaces and DIY repairs. A symbol of blue-collar craftsmanship.",
		ItemType:    ItemTypeMelee,
		BasePrice:   500,
		Damage:      5,
		MinRep:      RanksList[8].MinRep,
	},
	"crowbar": {
		Name:        "Crowbar",
		Description: "A versatile tool and makeshift weapon, the crowbar can pry open doors, crates, and skulls with equal efficiency.",
		ItemType:    ItemTypeMelee,
		BasePrice:   700,
		Damage:      7,
		MinRep:      RanksList[10].MinRep,
	},
	"switchblade": {
		Name:        "Switchblade",
		Description: "The switchblade, a compact folding knife, offers quick and discreet access for self-defense or utility in a pinch.",
		ItemType:    ItemTypeMelee,
		BasePrice:   900,
		Damage:      9,
		MinRep:      RanksList[12].MinRep,
	},
	"bbbat": {
		Name:        "Baseball Bat",
		Description: "A solid baseball bat, ideal for both sports and as an improvised weapon. Swing for the fences or fend off threats.",
		ItemType:    ItemTypeMelee,
		BasePrice:   1100,
		Damage:      11,
		MinRep:      RanksList[14].MinRep,
	},
	"fireaxe": {
		Name:        "Fire Axe",
		Description: "A heavy-duty fire axe, designed for breaking through obstacles and providing a powerful, life-saving tool in emergencies.",
		ItemType:    ItemTypeMelee,
		BasePrice:   1300,
		Damage:      13,
		MinRep:      RanksList[16].MinRep,
	},
	"machete": {
		Name:        "Machete",
		Description: "The machete, a rugged cutting tool, is essential for clearing foliage or defending against the wild and unexpected.",
		ItemType:    ItemTypeMelee,
		BasePrice:   1500,
		Damage:      15,
		MinRep:      RanksList[18].MinRep,
	},
	"katana": {
		Name:        "Katana",
		Description: "The katana, a legendary Japanese sword, balances precision and power, making it a deadly choice for any skilled warrior.",
		ItemType:    ItemTypeMelee,
		BasePrice:   1700,
		Damage:      17,
		MinRep:      RanksList[20].MinRep,
	},
	"chainsaw": {
		Name:        "Chainsaw",
		Description: "A fearsome chainsaw, designed for heavy-duty cutting and capable of unleashing raw, mechanical power in the right hands.",
		ItemType:    ItemTypeMelee,
		BasePrice:   1900,
		Damage:      19,
		MinRep:      RanksList[22].MinRep,
	},

	// Melee Armor
	"stabvest": {
		Name:        "Stab-Resistant Vest",
		Description: "Made of materials such as Kevlar and laminated fabrics. Designed to resist punctures from knives and other sharp-edged weapons.",
		ItemType:    ItemTypeArmor,
		BasePrice:   400,
		ArmorMelee:  4,
		MinRep:      RanksList[10].MinRep,
	},
	"chainmail": {
		Name:        "Chainmail",
		Description: "Modern chainmail is made from materials like steel or titanium rings. It provides good protection against slashing attacks from knives and swords.",
		ItemType:    ItemTypeArmor,
		BasePrice:   700,
		ArmorMelee:  7,
		MinRep:      RanksList[14].MinRep,
	},
	"hardarmor": {
		Name:        "Hard Plate Armor",
		Description: "Usually made from ceramics, steel, or composite materials. They provide excellent protection against knife and sword strikes",
		ItemType:    ItemTypeArmor,
		BasePrice:   1000,
		ArmorMelee:  10,
		MinRep:      RanksList[18].MinRep,
	},

	// Misc Items
	"smartphone": {
		Name:        "Smart Phone",
		Description: fmt.Sprintf("Use the get the location of druggies and dealer, for the small fee of $%d to your informant of cause.", settings.SmartPhoneCost),
		ItemType:    ItemTypeSmartPhone,
		BasePrice:   350,
		UseEffect:   "smartphone",
	},

	// NPC Equipment
	"deserteagle": {
		Name:        "Desert Eagle",
		Description: "The Desert Eagle .50 AE is a semi-automatic handgun that stands out for its considerable size and firepower. Chambered in .50 AE",
		ItemType:    ItemTypeGun,
		BasePrice:   500,
		Damage:      15,
	},
	"brokenbottle": {
		Name:        "Broken Bottle",
		Description: "A broken bottle that has been damaged, resulting in sharp and jagged edges.",
		ItemType:    ItemTypeMelee,
		BasePrice:   50,
		Damage:      2,
	},
	"guitar": {
		Name:        "Guitar",
		Description: "A six-stringed musical instrument known for its versatility and soulful melodies.",
		ItemType:    ItemTypeMelee,
		BasePrice:   100,
		Damage:      4,
	},
	"leadpipe": {
		Name:        "Lead Pipe",
		Description: "A heavy and sturdy cylindrical tube typically made of lead or other dense materials.",
		ItemType:    ItemTypeMelee,
		BasePrice:   100,
		Damage:      4,
	},
	"parcel": {
		Name:        "Parcel",
		Description: "A securely wrapped package or envelope intended for delivery or transport.",
		ItemType:    ItemTypeMelee,
		BasePrice:   50,
		Damage:      2,
	},
	"mac-10": {
		Name:        "MAC-10",
		Description: "\"Military Armament Corporation Model 10\" is a compact, blowback-operated submachine gun. Chambered in .45 ACP",
		ItemType:    ItemTypeGun,
		BasePrice:   400,
		Damage:      11,
	},
	"bikelock": {
		Name:        "Bike Lock",
		Description: "In the hands of a resourceful individual, it becomes an improvised weapon, offering reach and a blunt force impact",
		ItemType:    ItemTypeMelee,
		BasePrice:   75,
		Damage:      3,
	},

	// Trash / vendor Items
	"goldchain": {
		Name:        "Gold Chain",
		Description: "A accessory made of linked gold segments.",
		ItemType:    ItemTypeTrash,
		BasePrice:   300,
	},
	"festivalticket": {
		Name:        "Festival Ticket",
		Description: "A ticket for music festival",
		ItemType:    ItemTypeTrash,
		BasePrice:   150,
	},
	"sunglasses": {
		Name:        "Sun Glasses",
		Description: "Old, but decent brand.",
		ItemType:    ItemTypeTrash,
		BasePrice:   100,
	},
	"deliverypackage": {
		Name:        "\"Amazone\" Package",
		Description: "Something is inside.",
		ItemType:    ItemTypeTrash,
		BasePrice:   100,
	},
	"currentthing": {
		Name:        "Current Thing",
		Description: "An item which shows you support the \"Current Thing\".",
		ItemType:    ItemTypeTrash,
		BasePrice:   100,
	},
	"policebadge": {
		Name:        "Police Badge",
		Description: "Standard police badge.",
		ItemType:    ItemTypeTrash,
		BasePrice:   300,
	},
}
