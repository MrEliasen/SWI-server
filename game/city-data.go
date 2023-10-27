package game

// k = city/urban size in km2
// size = round(sqrt(k)) * 2
func NewCity(templ *CityTemplate) *City {
	return &City{
		Name:              templ.Name,
		ShortName:         templ.ShortName,
		ISO:               templ.ISO,
		Width:             uint(templ.Width),
		Height:            uint(templ.Height),
		TravelCostMin:     templ.TravelCostMin,
		TravelCostMax:     templ.TravelCostMax,
		BuildingLocations: templ.BuildingLocations,
		Grid:              make(map[string]*Location, templ.Height*templ.Width),
		NPCs:              make(map[NPCType]map[*Entity]bool),
		NpcSpawnList:      templ.NpcSpawnList,
		POILocations:      []Coordinates{},
		DrugDemands:       make(map[string]float32),
		Players:           make(map[*Client]bool),
		PlayerJoin:        make(chan *Client),
		PlayerLeave:       make(chan *Client),
		CityEvents:        make(chan bool),
	}
}

func GenerateCities() map[string]*City {
	cityList := map[string]*City{}

	for _, template := range CityTemplates {
		c := NewCity(&template)
		cityList[c.ShortName] = c
	}

	return cityList
}

type CityTemplate struct {
	Name              string
	ShortName         string
	ISO               string
	Width             uint8
	Height            uint8
	TravelCostMin     int64
	TravelCostMax     int64
	NpcSpawnList      map[NPCType]uint8
	BuildingLocations []BuildingLocation
}

var CityTemplates = []CityTemplate{
	{
		Name:          "Beijing",
		ShortName:     "BJ",
		Width:         30,
		Height:        30,
		TravelCostMin: 800,
		TravelCostMax: 1200,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 29,
					East:  20,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 12,
					East:  4,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 29,
					East:  21,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 6,
					East:  17,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 25,
					East:  10,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 8,
					East:  30,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Tokyo",
		ShortName:     "TY",
		Width:         30,
		Height:        30,
		TravelCostMin: 800,
		TravelCostMax: 1200,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 28,
					East:  4,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 12,
					East:  27,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 18,
					East:  8,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 24,
					East:  5,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 6,
					East:  15,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 9,
					East:  22,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Moscow",
		ShortName:     "MC",
		Width:         30,
		Height:        30,
		TravelCostMin: 240,
		TravelCostMax: 600,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 3,
					East:  29,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 16,
					East:  11,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 5,
					East:  19,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 23,
					East:  1,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 7,
					East:  21,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 28,
					East:  15,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Jakata",
		ShortName:     "JK",
		Width:         30,
		Height:        30,
		TravelCostMin: 900,
		TravelCostMax: 1300,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 20,
					East:  12,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 12,
					East:  27,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 8,
					East:  18,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 29,
					East:  3,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 14,
					East:  21,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 4,
					East:  7,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Mexico City",
		ShortName:     "MX",
		Width:         30,
		Height:        30,
		TravelCostMin: 300,
		TravelCostMax: 600,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 5,
					East:  18,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 14,
					East:  27,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 9,
					East:  8,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 20,
					East:  4,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 2,
					East:  25,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 15,
					East:  18,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "London",
		ShortName:     "LD",
		Width:         30,
		Height:        30,
		TravelCostMin: 600,
		TravelCostMax: 900,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 1,
					East:  24,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 23,
					East:  17,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 10,
					East:  6,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 29,
					East:  14,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 7,
					East:  29,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 25,
					East:  19,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Berlin",
		ShortName:     "BL",
		Width:         30,
		Height:        30,
		TravelCostMin: 200,
		TravelCostMax: 400,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 14,
					East:  26,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 23,
					East:  10,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 12,
					East:  26,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 5,
					East:  19,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 30,
					East:  7,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 8,
					East:  29,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Madrid",
		ShortName:     "MD",
		Width:         30,
		Height:        30,
		TravelCostMin: 600,
		TravelCostMax: 900,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 14,
					East:  8,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 12,
					East:  22,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 5,
					East:  17,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 30,
					East:  3,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 7,
					East:  10,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 14,
					East:  9,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Pretoria",
		ShortName:     "PT",
		Width:         30,
		Height:        30,
		TravelCostMin: 1200,
		TravelCostMax: 1800,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 24,
					East:  3,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 17,
					East:  12,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 28,
					East:  30,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 6,
					East:  3,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 19,
					East:  21,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 10,
					East:  29,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Rome",
		ShortName:     "RO",
		Width:         30,
		Height:        30,
		TravelCostMin: 200,
		TravelCostMax: 400,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 21,
					East:  18,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 22,
					East:  5,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 13,
					East:  18,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 7,
					East:  26,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 30,
					East:  2,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 25,
					East:  14,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Paris",
		ShortName:     "PA",
		Width:         30,
		Height:        30,
		TravelCostMin: 400,
		TravelCostMax: 900,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 19,
					East:  12,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 8,
					East:  27,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 16,
					East:  10,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 30,
					East:  15,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 4,
					East:  3,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 19,
					East:  22,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Warsaw",
		ShortName:     "WS",
		Width:         30,
		Height:        30,
		TravelCostMin: 200,
		TravelCostMax: 400,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 24,
					East:  7,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 11,
					East:  20,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 5,
					East:  9,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 28,
					East:  6,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 23,
					East:  12,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 2,
					East:  29,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Stockholm",
		ShortName:     "SH",
		Width:         30,
		Height:        30,
		TravelCostMin: 200,
		TravelCostMax: 400,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 27,
					East:  12,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 18,
					East:  7,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 1,
					East:  30,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 14,
					East:  23,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 10,
					East:  19,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 26,
					East:  2,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "New York City",
		ShortName:     "NY",
		Width:         30,
		Height:        30,
		TravelCostMin: 500,
		TravelCostMax: 900,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 25,
					East:  19,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 9,
					East:  16,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 21,
					East:  8,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 3,
					East:  27,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 24,
					East:  13,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 15,
					East:  12,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
	{
		Name:          "Las Vegas",
		ShortName:     "LV",
		Width:         30,
		Height:        30,
		TravelCostMin: 300,
		TravelCostMax: 600,
		NpcSpawnList: map[NPCType]uint8{
			DrugDealer:       2,
			DrugAddict:       2,
			Homeless:         6,
			Tweaker:          2,
			Bouncer:          2,
			Busker:           3,
			StreetVendor:     4,
			StreetGangMember: 2,
			Tourist:          5,
			Activist:         5,
			PoliceOfficer:    3,
			BeatCop:          2,
			DeliveryDriver:   4,
		},
		BuildingLocations: []BuildingLocation{
			{
				Coords: Coordinates{
					North: 30,
					East:  13,
				},
				Buildings: []BuildingType{
					BuildingTypePawnShop,
				},
			},
			{
				Coords: Coordinates{
					North: 12,
					East:  25,
				},
				Buildings: []BuildingType{
					BuildingTypeAirport,
				},
			},
			{
				Coords: Coordinates{
					North: 17,
					East:  11,
				},
				Buildings: []BuildingType{
					BuildingTypeHospital,
				},
			},
			{
				Coords: Coordinates{
					North: 29,
					East:  6,
				},
				Buildings: []BuildingType{
					BuildingTypeArms,
				},
			},
			{
				Coords: Coordinates{
					North: 4,
					East:  21,
				},
				Buildings: []BuildingType{
					BuildingTypeBank,
				},
			},
			{
				Coords: Coordinates{
					North: 7,
					East:  14,
				},
				Buildings: []BuildingType{
					BuildingTypeBar,
				},
			},
		},
	},
}
