package main

import (
	"flag"

	"metrics/internal/server/server"
)

func main() {
	var httpServer server.Server
	flag.Parse()

	httpServer.Run()
}
