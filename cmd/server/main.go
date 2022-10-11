// HTTP сервер для runtime метрик
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"metrics/internal/server/config"
	"metrics/internal/server/server"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func Profiling(addr string) {
	log.Fatal(http.ListenAndServe(addr, nil))
}

func main() {
	ctx := context.Background()

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	appConfig := config.LoadConfig()
	appServer := server.NewServer(appConfig)

	if appServer.Config().ProfilingAddr != "" {
		go Profiling(appServer.Config().ProfilingAddr)
	}

	err := appServer.Run(ctx)
	if err != nil {
		log.Println(err)
	}
}
