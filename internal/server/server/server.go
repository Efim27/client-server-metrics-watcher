package server

import (
	"log"
	"net/http"
	"time"

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

	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.GzipHandle)

	//json handler
	router.Post("/value/", server.JSONStatValue)

	router.Get("/", server.PrintStatsValues)
	router.Get("/ping", server.PingGet)
	router.Get("/value/{statType}/{statName}", server.PrintStatValue)

	router.Route("/update/", func(router chi.Router) {
		//json handler
		router.Post("/", server.UpdateStatJSONPost)

		router.Post("/gauge/{statName}/{statValue}", server.UpdateGaugePost)
		router.Post("/counter/{statName}/{statValue}", server.UpdateCounterPost)
		router.Post("/{statType}/{statName}/{statValue}", server.UpdateNotImplementedPost)
	})

	server.chiRouter = router
}

func (server *Server) Run() {
	server.initStorage()
	defer server.storage.Close()
	server.initRouter()

	log.Fatal(http.ListenAndServe(server.config.ServerAddr, server.chiRouter))
}
