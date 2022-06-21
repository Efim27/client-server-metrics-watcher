package main

import (
	"metrics/internal/agent"
	"metrics/internal/agent/config"
)

func main() {
	config := config.LoadConfig()
	app := agent.NewHTTPClient(config)
	app.Run()
}
