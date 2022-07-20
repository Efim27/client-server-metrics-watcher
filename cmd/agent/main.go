package main

import (
	"metrics/internal/agent"
	"metrics/internal/agent/config"
)

func main() {
	appConfig := config.LoadConfig()
	app := agent.NewAppHTTP(appConfig)
	app.Run()
}
