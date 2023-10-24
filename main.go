package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mreliasen/swi-server/game"
	"github.com/mreliasen/swi-server/internal/database"
	"github.com/mreliasen/swi-server/internal/logger"
)

var (
	env    = flag.String("env", "prod", "Environment")
	domain = flag.String("domain", "swi-server.sirmre.com", "server domain (TLS)")
	dburl  = flag.String("dburl", "ws://127.0.0.1:8080", "DB Url/path")
)

func CORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		h(w, r)
	}
}

func main() {
	flag.Parse()
	logger.New(env)
	logger.Logger.Info("Server starting..")

	ctx, cancel := context.WithCancel(context.Background())

	gracefulShutdown := make(chan os.Signal, 1)
	shutdownSave := make(chan int, 1)
	signal.Notify(gracefulShutdown, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	db, ok := database.Connect(*dburl)
	if !ok {
		os.Exit(1)
		return
	}

	defer db.Close()

	gameInstance := game.NewGame(db)
	go gameInstance.Run()

	logger.Logger.Info("Game Ready!")

	go StartWebServer(domain, gameInstance)

	if strings.ToLower(*env) == "prod" {
		go gameInstance.RenderConsoleUI()
	}

	<-gracefulShutdown
	close(gracefulShutdown)

	go func() {
		logger.Logger.Warn("Saving player data..")
		gameInstance.Save()
		logger.Logger.Warn("Saving Complete")
		shutdownSave <- 1
	}()

	<-shutdownSave
	close(shutdownSave)

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Fatal(fmt.Sprintf("HTTP sever shutdown error: %s", err))
		defer os.Exit(1)
	} else {
		logger.Logger.Info("Server gracefully stopped.")
	}

	cancel()
	defer os.Exit(0)
}
