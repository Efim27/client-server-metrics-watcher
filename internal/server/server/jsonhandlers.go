package server

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	"metrics/internal/server/responses"
	"metrics/internal/server/storage"
)

// UpdateMetricPostJSON
// @Tags Update
// @Summary Update metric value
// @ID updateMetricPostJSON
// @Produce json
// @Success 200
// @Failure 400
// @Router /update/ [post]
func (server Server) UpdateMetricPostJSON(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	inputJSON := struct {
		storage.Metric
		Hash string `json:"hash,omitempty"`
	}{}
	response := responses.NewUpdateMetricResponse()

	//JSON decoding
	err := json.NewDecoder(request.Body).Decode(&inputJSON)
	if err != nil {
		http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
		return
	}

	//Validation
	_, err = govalidator.ValidateStruct(inputJSON)
	if err != nil {
		http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
		return
	}

	newMetricValue := storage.MetricValue{
		MType: inputJSON.MType,
		Value: inputJSON.Value,
		Delta: inputJSON.Delta,
	}

	//Check sign
	var metricHash []byte
	if server.config.SignKey != "" {
		var requestMetricHash []byte
		requestMetricHash, err = hex.DecodeString(inputJSON.Hash)
		if err != nil {
			http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
			return
		}

		metricHash = newMetricValue.GetHash(inputJSON.ID, server.config.SignKey)
		if !hmac.Equal(requestMetricHash, metricHash) {
			http.Error(rw, response.SetStatusError(errors.New("invalid hash")).GetJSONString(), http.StatusBadRequest)
			return
		}
	}

	//Update value
	err = server.storage.Update(inputJSON.ID, newMetricValue)
	if err != nil {
		http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(response.SetHash(hex.EncodeToString(metricHash)).GetJSONBytes())
}

// UpdateMetricBatchJSON
// @Tags Update
// @Summary Update metric value using batch JSON
// @ID updateMetricBatchJSON
// @Produce json
// @Param JSON body []storage.Metric true "JSON"
// @Success 200
// @Failure 400
// @Router /updates/ [post]
func (server Server) UpdateMetricBatchJSON(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	var MetricBatch []storage.Metric
	response := responses.NewUpdateMetricResponse()

	//JSON decoding
	err := json.NewDecoder(request.Body).Decode(&MetricBatch)
	if err != nil {
		http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
		return
	}

	//Validation
	for _, OneMetric := range MetricBatch {
		_, err = govalidator.ValidateStruct(OneMetric)
		if err != nil {
			http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
			return
		}
	}

	err = server.storage.UpdateManySliceMetric(MetricBatch)
	if err != nil {
		http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(response.GetJSONBytes())
}

// MetricValuePostJSON
// @Tags Value
// @Summary Metric value JSON
// @ID metricValuePostJSON
// @Produce json
// @Param JSON body handlers.InputMetricsJSON true "JSON"
// @Success 200
// @Failure 400
// @Failure 404
// @Router /value/ [post]
func (server Server) MetricValuePostJSON(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	var InputMetricsJSON struct {
		ID    string `json:"id" valid:"required"`
		MType string `json:"type" valid:"required,in(counter|gauge)"`
	}

	err := json.NewDecoder(request.Body).Decode(&InputMetricsJSON)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = govalidator.ValidateStruct(InputMetricsJSON)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	statValue, err := server.storage.Read(InputMetricsJSON.ID, InputMetricsJSON.MType)
	if err != nil {
		http.Error(rw, "Unknown statName", http.StatusNotFound)
		return
	}

	answerJSON := struct {
		storage.Metric
		Hash string `json:"hash"`
	}{
		Metric: storage.Metric{
			ID: InputMetricsJSON.ID,
			MetricValue: storage.MetricValue{
				MType: statValue.MType,
				Delta: statValue.Delta,
				Value: statValue.Value,
			},
		},
	}

	if server.config.SignKey != "" {
		answerJSON.Hash = hex.EncodeToString(answerJSON.GetHash(InputMetricsJSON.ID, server.config.SignKey))
	}

	rw.WriteHeader(http.StatusOK)
	err = json.NewEncoder(rw).Encode(answerJSON)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

// PingGetJSON
// @Summary Ping
// @ID pingGetJSON
// @Produce json
// @Success 200
// @Failure 500
// @Router /ping [get]
func (server Server) PingGetJSON(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response := responses.NewDefaultResponse()
	pingError := server.storage.Ping()

	if pingError != nil {
		http.Error(rw, response.SetStatusError(pingError).GetJSONString(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(response.GetJSONBytes())
}
