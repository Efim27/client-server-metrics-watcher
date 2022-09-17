package main

import (
	"metrics/internal/server/config"
	"metrics/internal/server/server"
)

func main() {
	appConfig := config.LoadConfig()
	appServer := server.NewServer(appConfig)
	appServer.Run()
}
