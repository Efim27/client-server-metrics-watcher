// Package server - обработчики сервера
//
// Подробнее в swagger документации
package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"metrics/internal/server/storage"
)

// UpdateGaugePost
// @Tags Update
// @Summary Update gauge metric
// @ID updateGaugePost
// @Produce plain
// @Param statName query string false "Имя метрики"
// @Param statValue query string false "Значение"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /update/gauge/{statName}/{statValue} [post]
func (server Server) UpdateGaugePost(rw http.ResponseWriter, request *http.Request) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statValueFloat, err := strconv.ParseFloat(statValue, 64)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Bad request"))
		return
	}

	err = server.storage.Update(statName, storage.MetricValue{
		MType: storage.MeticTypeGauge,
		Value: &statValueFloat,
	})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Server error"))
		return
	}

	log.Println("Update gauge:")
	log.Printf("%v: %v\n", statName, statValue)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Ok"))
}

// UpdateCounterPost
// @Tags Update
// @Summary Update counter metric
// @ID updateCounterPost
// @Produce plain
// @Param statName query string false "Имя метрики"
// @Param statValue query string false "Значение"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /update/counter/{statName}/{statValue} [post]
func (server Server) UpdateCounterPost(rw http.ResponseWriter, request *http.Request) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statValueInt, err := strconv.ParseInt(statValue, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	err = server.storage.Update(statName, storage.MetricValue{
		MType: storage.MeticTypeCounter,
		Delta: &statValueInt,
	})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	log.Println("Inc counter:")
	log.Printf("%v: %v\n", statName, statValue)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Ok"))
}

// UpdateNotImplementedPost
// @Tags Update
// @Summary Update not implemented
// @ID updateNotImplementedPost
// @Produce plain
// @Param statType query string false "Тип метрики" Enums(gauge, counter) default(gauge)
// @Param statName query string false "Имя метрики"
// @Param statValue query string false "Значение"
// @Failure 501
// @Router /update/{statType}/{statName}/{statValue} [post]
func (server Server) UpdateNotImplementedPost(rw http.ResponseWriter, _ *http.Request) {
	log.Println("Update not implemented statType")

	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not implemented"))
}

// PrintMetricGet
// @Tags Value
// @Summary Metric value
// @ID printMetricGet
// @Produce plain
// @Param statType query string false "Тип метрики" Enums(gauge, counter) default(gauge)
// @Param statName query string false "Имя метрики"
// @Success 200
// @Failure 404
// @Router /value/{statType}/{statName} [get]
func (server Server) PrintMetricGet(rw http.ResponseWriter, request *http.Request) {
	statType := chi.URLParam(request, "statType")
	statName := chi.URLParam(request, "statName")

	metric, err := server.storage.Read(statName, statType)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Unknown statName"))
		return
	}

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(metric.GetStringValue()))
}
