package main

import (
	"librus/pkg/logger"
	"librus/telegram"
)

func main() {
	// Initialize logger first
	logger.Initialize()
	defer logger.Sync()

	logger.Info("Starting Librus Telegram Bot")
	telegram.Start()
}
