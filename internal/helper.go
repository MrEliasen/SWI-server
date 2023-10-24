package internal

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/mreliasen/swi-server/internal/logger"
)

var (
	memory      uint32 = 64 * 1024
	iterations  uint8  = 3
	parallelism uint32 = 2
	saltLength  uint32 = 16
	keyLength   uint32 = 32
)

var IsValidCharacterName = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,15}$`).MatchString

func PrintStruct(s any) {
	fmt.Printf("%+v\n", s)
}

func HashPassword(password string) (string, error) {
	params := &argon2id.Params{
		Memory:      memory,
		Iterations:  uint32(iterations),
		Parallelism: uint8(parallelism),
		SaltLength:  saltLength,
		KeyLength:   keyLength,
	}

	hash, err := argon2id.CreateHash(password, params)
	if err != nil {
		logger.Logger.Fatal("Failed to hash password")
		log.Fatal(err)
	}

	return hash, nil
}

func CheckPassword(password string, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Fatal(err)
	}

	return match, nil
}

func ToTable(headings []string, data [][]string) []string {
	longest := []int{}
	padding := []string{}
	spacer := []string{}

	for _, h := range headings {
		longest = append(longest, len(h))
		spacer = append(spacer, "")
	}

	tableLines := [][]string{
		spacer,
		headings,
		spacer,
	}

	for _, d := range data {
		row := []string{}
		for i, c := range d {
			l := len(c)

			if l > longest[i] {
				longest[i] = l
			}

			row = append(row, c)
		}

		tableLines = append(tableLines, row)
	}

	for _, d := range longest {
		padding = append(padding, "%-"+fmt.Sprintf("%d", d+1)+"s")
	}

	lineLng := 0
	result := []string{}

	for _, line := range tableLines {
		for i, cell := range line {
			pd := padding[i]
			line[i] = fmt.Sprintf(pd, cell)
		}

		str := "| " + strings.Join(line, "| ") + " |"

		if len(str) > lineLng {
			lineLng = len(str)
		}

		result = append(result, str)
	}

	separator := strings.Repeat("-", lineLng)
	result[0] = separator
	result[2] = separator
	result = append(result, separator)
	return result
}
