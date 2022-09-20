package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"metrics/internal/logger"
	"metrics/internal/server/config"
	"metrics/internal/server/middleware"
	"metrics/internal/server/storage"
)

type Server struct {
	logFile   *os.File
	logger    *zap.Logger
	storage   storage.MetricStorager
	chiRouter chi.Router
	config    config.Config
	startTime time.Time
}

func NewServer(config config.Config) *Server {
	log.Println(111)
	server := &Server{
		config: config,
	}
	log.Println(config)
	mustInitLogger(server)
	server.logger.Info("load config successfully", zap.Any("config", server.config))

	return server
}

func mustInitLogger(server *Server) {
	log.Println(222)
	logLevel := zap.InfoLevel
	if server.config.DebugMode {
		logLevel = zap.DebugLevel
	}

	if server.config.LogFile != "" {
		logFile, err := os.OpenFile(server.config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("cant open log file")
		}

		server.logFile = logFile
	}

	server.logger = logger.InitializeLogger(server.logFile, logLevel)
	server.logger.Info("debug mode enabled")
}

func (server *Server) selectStorage() storage.MetricStorager {
	storageConfig := server.config.Store

	if storageConfig.DatabaseDSN != "" {
		server.logger.Info("use database storage")
		repository, err := storage.NewDBRepo(storageConfig)
		if err != nil {
			panic(err)
		}

		return repository
	}

	server.logger.Info("use memory storage")
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
	log.Println(333)
	server.logger.Info("start")

	server.initStorage()
	server.logger.Info("init storage successfully")
	defer server.storage.Close()

	server.initRouter()
	serverHTTP := &http.Server{
		Addr:    server.config.ServerAddr,
		Handler: server.chiRouter,
	}

	go func() {
		err := serverHTTP.ListenAndServe()
		if err != nil {
			server.logger.Info("HTTP server closed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	err := serverHTTP.Close()
	if err != nil {
		server.logger.Fatal("HTTP server stop error", zap.Error(err))
	}
}

func (server *Server) Config() (config config.Config) {
	return server.config
}
