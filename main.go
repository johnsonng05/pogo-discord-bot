package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"pogo-bot/internal/bot"
	"pogo-bot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}
	defer b.Close()

	if err := b.Start(); err != nil {
		log.Fatalf("start: %v", err)
	}

	log.Println("pogo-bot is running. Press Ctrl+C to exit.")

	// Block until SIGINT / SIGTERM so the process stays alive for the
	// websocket and the background ticker.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down")
}
