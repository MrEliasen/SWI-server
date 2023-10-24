package tools

import (
	"fmt"
	"math"
	"strings"

	"github.com/pterm/pterm"
)

var (
	base_xp       = 100.0
	growth_factor = 3.0
	Ranks         = []string{
		"Crackhead",
		"Street Trash",
		"Disrespectable Punk",
		"Nobody",
		"Wannabe",
		"Slacker",
		"Street Punk",
		"Thug Wannabe",
		"Thug",
		"Hustler",
		"Wanskta",
		"Gangster",
		"Soldier",
		"Playa",
		"Pimp",
		"Pusher",
		"Smuggler",
		"Gun Runner",
		"Mobster",
		"Drug Lord",
		"Capo",
		"Underboss",
		"Don",
		"Kingpin",
		"Godfather",
		"Most Wanted",
		"Worlds Most Wanted",
	}
)

func calculateExperienceRequired(level int) uint64 {
	ranks := len(Ranks)
	levelDiscount := (float64(ranks) - float64(level)) / 100
	expNeeded := math.Round(base_xp * math.Pow(float64(level), growth_factor))

	if level > 1 && levelDiscount > 0 {
		discountPercent := 1.0 - levelDiscount*2
		expNeeded = expNeeded * discountPercent
	}

	return uint64(expNeeded/100) * 100
}

func GenerateExperienceTable() {
	skip := 3
	tableData := pterm.TableData{
		{"Rank", "Exp Required"},
	}
	barData := []pterm.Bar{}

	code := []string{
		"switch {",
	}
	code2 := []string{}

	for i := 0 + skip; i < len(Ranks)-skip; i++ {
		exp := calculateExperienceRequired(i - skip)
		code = append(code, fmt.Sprintf("case rep <= %d:", exp))
		code = append(code, fmt.Sprintf("return \"%s\"", Ranks[i]))
		code2 = append(code2, fmt.Sprintf("{Name: \"%s\", Value: %d},", Ranks[i], exp))
		tableData = append(tableData, []string{
			Ranks[i],
			fmt.Sprintf("%d", exp),
		})
		barData = append(barData, pterm.Bar{Label: fmt.Sprintf("%d", i), Value: int(exp)})
	}

	code = append(code, "default:")
	code = append(code, "return \"Enigma\"")
	code = append(code, "}")
	code = append(code, "")
	code2 = append(code2, "")

	pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	pterm.DefaultBarChart.WithBars(barData).WithHorizontal().WithShowValue().Render()

	area, _ := pterm.DefaultArea.Start()
	area.Update(strings.Join(code, "\n") + "\n\n\n" + strings.Join(code2, "\n"))
	area.Stop()
}
