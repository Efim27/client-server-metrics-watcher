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
	startTime time.Time
	chiRouter chi.Router
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

func (server *Server) Run() {
	memStatsStorage := storage.NewMemStatsMemoryRepo()
	defer memStatsStorage.Close()
	server.chiRouter = newRouter(memStatsStorage)

	log.Fatal(http.ListenAndServe(config.AppConfig.ServerAddr, server.chiRouter))
}
