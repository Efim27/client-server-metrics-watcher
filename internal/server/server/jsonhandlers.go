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
		requestMetricHash, err := hex.DecodeString(inputJSON.Hash)
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

func (server Server) MetricValuePostJSON(rw http.ResponseWriter, request *http.Request) {
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
		answerJSON.Hash = hex.EncodeToString(answerJSON.Metric.GetHash(InputMetricsJSON.ID, server.config.SignKey))
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	err = json.NewEncoder(rw).Encode(answerJSON)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (server Server) PingGetJSON(rw http.ResponseWriter, request *http.Request) {
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
