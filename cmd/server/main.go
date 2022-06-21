package main

import (
	"metrics/internal/server/config"
	"metrics/internal/server/server"
)

func main() {
	config := config.LoadConfig()
	server := server.NewServer(config)
	server.Run()
}
