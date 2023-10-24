package game

type Rank struct {
	Name     string `json:"name"`
	MinRep   int64  `json:"rep"`
	NextRank *Rank
	PrevRank *Rank
}

var RanksList = [...]*Rank{
	{Name: "Crackhead", MinRep: -1400},
	{Name: "Street Trash", MinRep: -400},
	{Name: "Disrespectable Punk", MinRep: -100},
	{Name: "Nobody", MinRep: 0},
	{Name: "Wannabe", MinRep: 100},
	{Name: "Slacker", MinRep: 400},
	{Name: "Street Punk", MinRep: 1400},
	{Name: "Thug Wannabe", MinRep: 3400},
	{Name: "Thug", MinRep: 7000},
	{Name: "Hustler", MinRep: 12500},
	{Name: "Wanskta", MinRep: 20500},
	{Name: "Gangster", MinRep: 31700},
	{Name: "Soldier", MinRep: 46600},
	{Name: "Playa", MinRep: 65900},
	{Name: "Pimp", MinRep: 90500},
	{Name: "Pusher", MinRep: 120900},
	{Name: "Smuggler", MinRep: 158100},
	{Name: "Gun Runner", MinRep: 203000},
	{Name: "Mobster", MinRep: 256500},
	{Name: "Drug Lord", MinRep: 319400},
	{Name: "Capo", MinRep: 393000},
	{Name: "Underboss", MinRep: 478200},
	{Name: "Don", MinRep: 576100},
	{Name: "Kingpin", MinRep: 688000},
}

func LinkRanks() {
	total := len(RanksList) - 1
	for i, r := range RanksList {
		if i > 0 {
			r.PrevRank = RanksList[i-1]
		}

		if i < total {
			r.NextRank = RanksList[i+1]
		}
	}
}

func GetRank(rep int64) *Rank {
	for _, r := range RanksList {
		if rep <= r.MinRep {
			return r
		}
	}

	return &Rank{
		Name:   "Enigma",
		MinRep: 0,
	}
}
