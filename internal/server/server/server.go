package server

import (
	"context"
	"log"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"metrics/internal/server/config"
	"metrics/internal/server/middleware"
	"metrics/internal/server/storage"
)

type Server struct {
	storage   storage.MetricStorager
	chiRouter chi.Router
	config    config.Config
	startTime time.Time
}

func NewServer(config config.Config) *Server {
	log.Println(config)

	return &Server{
		config: config,
	}
}

func (server *Server) selectStorage() storage.MetricStorager {
	storageConfig := server.config.Store

	if storageConfig.DatabaseDSN != "" {
		log.Println("DB Storage")
		repository, err := storage.NewDBRepo(storageConfig)
		if err != nil {
			panic(err)
		}

		return repository
	}

	log.Println("Memory Storage")
	repository := storage.NewMetricsMemoryRepo(storageConfig)

	return repository
}

func (server *Server) initStorage() {
	metricsMemoryRepo := server.selectStorage()
	server.storage = metricsMemoryRepo

	if server.config.Store.Restore {
		server.storage.InitFromFile()
	}
}

func (server *Server) initRouter() {
	router := chi.NewRouter()

	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.GzipHandle)

	router.Get("/", server.PrintAllMetricStatic)
	router.Get("/ping", server.PingGetJSON)
	router.Get("/value/{statType}/{statName}", server.PrintMetricGet)

	router.Post("/value/", server.MetricValuePostJSON)
	router.Post("/updates/", server.UpdateMetricBatchJSON)

	router.Route("/update/", func(router chi.Router) {
		router.Post("/", server.UpdateMetricPostJSON)

		router.Post("/gauge/{statName}/{statValue}", server.UpdateGaugePost)
		router.Post("/counter/{statName}/{statValue}", server.UpdateCounterPost)
		router.Post("/{statType}/{statName}/{statValue}", server.UpdateNotImplementedPost)
	})

	server.chiRouter = router
}

func (server *Server) Run(ctx context.Context) {
	server.initStorage()
	defer server.storage.Close()

	server.initRouter()
	serverHTTP := &http.Server{
		Addr:    server.config.ServerAddr,
		Handler: server.chiRouter,
	}

	go func() {
		err := serverHTTP.ListenAndServeTLS("./keysSSL/server.crt", "./keysSSL/server.key")
		if err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()
	err := serverHTTP.Close()
	if err != nil {
		log.Println(err)
	}
}

func (server *Server) Config() (config config.Config) {
	return server.config
}
