package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"metrics/internal/server/config"
	"metrics/internal/server/server"
)

func Profiling(addr string) {
	log.Fatal(http.ListenAndServe(addr, nil))
}

func main() {
	appConfig := config.LoadConfig()
	appServer := server.NewServer(appConfig)

	go Profiling(appServer.Config().ProfilingAddr)
	appServer.Run()
}
