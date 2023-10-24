package logger

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

var (
	startTime       = time.Now()
	Logger          *pterm.Logger
	logTransactions *log.Logger
	logCombat       *log.Logger
	logItems        *log.Logger
	logChat         *log.Logger
)

func LogItems(name string, action string, item string, north int, east int, city string) {
	logItems.Printf("%s, %s, %s, %s N%d-E%d", name, action, item, city, north, east)
}

func LogBuySell(name string, action string, amount uint32, item string) {
	logTransactions.Printf(",%s,%s,%d,%s", name, action, amount, item)
}

func LogCombat(attacker string, victim string, action string, weapon string, damage uint, north int, east int, city string) {
	logCombat.Printf(",%s,%s,%s,%s,%d,%s N%d-E%d", attacker, action, victim, weapon, damage, city, north, east)
}

func LogChat(msgType string, name string, message string, recipient string) {
	logChat.Printf(",%s,%s,%s,%s", msgType, name, message, recipient)
}

func New(env *string) {
	logLevel := pterm.LogLevelWarn

	if strings.ToLower(*env) != "prod" {
		logLevel = pterm.LogLevelTrace
	}

	logger := *pterm.DefaultLogger.WithLevel(logLevel)

	logsDirPath, err := filepath.Abs("./logs")
	os.MkdirAll(logsDirPath, os.ModePerm)

	if err != nil {
		fmt.Println("Error reading logs directory path:", err)
	}

	transactions, err := os.OpenFile(logsDirPath+"/transactions.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println("Error opening transactions.log:", err)
		os.Exit(1)
	}

	combat, err := os.OpenFile(logsDirPath+"/combat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println("Error opening combat.log:", err)
		os.Exit(1)
	}

	items, err := os.OpenFile(logsDirPath+"/items.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println("Error opening items.log:", err)
		os.Exit(1)
	}

	chat, err := os.OpenFile(logsDirPath+"/chat.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println("Error opening chat.log:", err)
		os.Exit(1)
	}

	logTransactions = log.New(transactions, "Transactions ", log.LstdFlags)
	logCombat = log.New(combat, "Combat", log.LstdFlags)
	logItems = log.New(items, "Items", log.LstdFlags)
	logChat = log.New(chat, "Chat", log.LstdFlags)
	Logger = &logger
}

func Uptime() string {
	now := time.Now()

	runTimeDiff := now.Sub(startTime)
	days := math.Floor(runTimeDiff.Hours() / 24)
	hours := math.Floor(runTimeDiff.Hours() - days*24)
	minutes := math.Floor(runTimeDiff.Minutes() - hours*24)
	seconds := runTimeDiff.Seconds() - minutes*60

	return fmt.Sprintf("Uptime: %d days, %d hours, %d minutes, %d seconds.", int64(days), int64(hours), int64(minutes), int64(seconds))
}
