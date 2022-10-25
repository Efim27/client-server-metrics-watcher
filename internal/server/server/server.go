package server

import (
	"context"
	"crypto/rsa"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	handlerRSA "metrics/internal/rsa"
	grpcServices "metrics/internal/server/grpc"

	_ "net/http/pprof"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"metrics/internal/server/config"
	"metrics/internal/server/middleware"
	"metrics/internal/server/storage"
	pb "metrics/proto"
)

type Server struct {
	storage       storage.MetricStorager
	chiRouter     chi.Router
	config        config.Config
	privateKeyRSA *rsa.PrivateKey
	startTime     time.Time
	serverGRPC    *grpc.Server
}

func NewServer(config config.Config) (server *Server) {
	var err error

	server = &Server{
		config:     config,
		serverGRPC: grpc.NewServer(),
	}
	log.Println(server.config)

	if config.PrivateKeyRSA != "" {
		server.privateKeyRSA, err = handlerRSA.ParsePrivateKeyRSA(config.PrivateKeyRSA)
	}
	if err != nil {
		log.Fatal("Parsing RSA key error")
	}

	return
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

	if server.privateKeyRSA != nil {
		RSAHandle := middleware.NewRSAHandle(server.privateKeyRSA)
		router.Use(RSAHandle)
	}

	if server.config.TrustedSubNet != "" {
		SubNetHandle := middleware.NewSubNetHandle(server.config.TrustedSubNet)
		router.Use(SubNetHandle)
	}

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

func (server *Server) RunServerGRPC() (err error) {
	lis, err := net.Listen("tcp", server.config.ServerGRPCAddr)
	if err != nil {
		return
	}

	pb.RegisterMetricsServer(server.serverGRPC, grpcServices.NewMetricsService(server.storage))

	go func() {
		err = server.serverGRPC.Serve(lis)
		if err != nil {
			log.Println(err)
		}
	}()

	return
}

func (server *Server) Run(ctx context.Context) (err error) {
	server.initStorage()
	defer server.storage.Close()

	server.initRouter()
	serverHTTP := &http.Server{
		Addr:    server.config.ServerAddr,
		Handler: server.chiRouter,
	}

	eventServerStopped := sync.WaitGroup{}
	eventServerStopped.Add(1)
	go func() {
		<-ctx.Done()

		if err = serverHTTP.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		server.serverGRPC.GracefulStop()

		if server.config.Store.Interval != storage.SyncUploadSymbol {
			err = server.storage.Save()
			if err != nil {
				log.Println(err)
			}
		}

		eventServerStopped.Done()
	}()

	if server.config.ServerGRPCAddr != "" {
		err = server.RunServerGRPC()
		if err != nil {
			log.Fatal(err)
		}
	}

	err = serverHTTP.ListenAndServeTLS("./keysSSL/server.crt", "./keysSSL/server.key")
	if errors.Is(err, fs.ErrNotExist) {
		log.Println("SSL keys not found, using HTTP")
		err = serverHTTP.ListenAndServe()
	}
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Server stopping...")
		eventServerStopped.Wait()
		log.Println("Server stopped successfully")
	}

	return
}

func (server *Server) Config() (config config.Config) {
	return server.config
}
