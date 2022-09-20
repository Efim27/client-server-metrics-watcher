// HTTP сервер для runtime метрик
package main

import (
	"context"
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
	ctx := context.Background()

	log.Println(1111)
	appConfig := config.LoadConfig()
	log.Println(2222)
	appServer := server.NewServer(appConfig)
	log.Println(3333)

	if appServer.Config().ProfilingAddr != "" {
		log.Println(4444)
		go Profiling(appServer.Config().ProfilingAddr)
	}

	log.Println(5555)
	appServer.Run(ctx)
}
