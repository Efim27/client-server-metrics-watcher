package server

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"metrics/internal/server/config"
	"metrics/internal/server/handlers"
	"metrics/internal/server/storage"
)

type Server struct {
	chiRouter chi.Router
	config    config.Config
	startTime time.Time
}

func newRouter(memStatsStorage storage.MemStatsMemoryRepo) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	//Маршруты
	router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		handlers.PrintStatsValues(writer, request, memStatsStorage)
	})

	//json handler
	router.Post("/value/", func(writer http.ResponseWriter, request *http.Request) {
		handlers.JSONStatValue(writer, request, memStatsStorage)
	})

	router.Get("/value/{statType}/{statName}", func(writer http.ResponseWriter, request *http.Request) {
		handlers.PrintStatValue(writer, request, memStatsStorage)
	})

	router.Route("/update/", func(router chi.Router) {
		//json handler
		router.Post("/", func(writer http.ResponseWriter, request *http.Request) {
			handlers.UpdateStatJSONPost(writer, request, memStatsStorage)
		})

		router.Post("/gauge/{statName}/{statValue}", func(writer http.ResponseWriter, request *http.Request) {
			handlers.UpdateGaugePost(writer, request, memStatsStorage)
		})
		router.Post("/counter/{statName}/{statValue}", func(writer http.ResponseWriter, request *http.Request) {
			handlers.UpdateCounterPost(writer, request, memStatsStorage)
		})
		router.Post("/{statType}/{statName}/{statValue}", func(writer http.ResponseWriter, request *http.Request) {
			handlers.UpdateNotImplementedPost(writer, request)
		})
	})

	return router
}

func NewServer(config config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (server *Server) Run() {
	memStatsStorage := storage.NewMemStatsMemoryRepo(server.config.Store)
	defer memStatsStorage.Close()
	if server.config.Store.Restore {
		memStatsStorage.InitFromFile()
	}
	server.chiRouter = newRouter(memStatsStorage)
	log.Println(server.config)

	log.Fatal(http.ListenAndServe(server.config.ServerAddr, server.chiRouter))
}
