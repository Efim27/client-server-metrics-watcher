package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"metrics/internal/server/storage"
)

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

	server.logger.Debug("update gauge metric", zap.String("name", statName), zap.String("value", statValue))
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Ok"))
}

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

	server.logger.Debug("increment counter metric", zap.String("name", statName), zap.String("value", statValue))
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Ok"))
}

func (server Server) UpdateNotImplementedPost(rw http.ResponseWriter, _ *http.Request) {
	server.logger.Debug("update not implemented statType")

	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not implemented"))
}

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
