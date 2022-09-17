package main

import (
	"metrics/internal/agent"
	"metrics/internal/agent/config"
)

func main() {
	appConfig := config.LoadConfig()
	app := agent.NewHTTPClient(appConfig)
	app.Run()
}
