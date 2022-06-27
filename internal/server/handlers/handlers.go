package handlers

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi"
	"metrics/internal/server/responses"
	"metrics/internal/server/storage"
)

//UpdateStatJSONPost update stat via json
func UpdateStatJSONPost(rw http.ResponseWriter, request *http.Request, metricsMemoryRepo storage.MetricStorager, SignKey string) {
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
	if SignKey != "" {
		requestMetricHash, err := hex.DecodeString(inputJSON.Hash)
		if err != nil {
			http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
			return
		}

		metricHash = newMetricValue.GetHash(inputJSON.ID, SignKey)
		if !hmac.Equal(requestMetricHash, metricHash) {
			http.Error(rw, response.SetStatusError(errors.New("invalid hash")).GetJSONString(), http.StatusBadRequest)
			return
		}
	}

	//Update value
	err = metricsMemoryRepo.Update(inputJSON.ID, newMetricValue)
	if err != nil {
		http.Error(rw, response.SetStatusError(err).GetJSONString(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(response.SetHash(hex.EncodeToString(metricHash)).GetJSONBytes())
}

func UpdateGaugePost(rw http.ResponseWriter, request *http.Request, metricsMemoryRepo storage.MetricStorager) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statValueFloat, err := strconv.ParseFloat(statValue, 64)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Bad request"))
		return
	}

	err = metricsMemoryRepo.Update(statName, storage.MetricValue{
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

func UpdateCounterPost(rw http.ResponseWriter, request *http.Request, metricsMemoryRepo storage.MetricStorager) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statValueInt, err := strconv.ParseInt(statValue, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	err = metricsMemoryRepo.Update(statName, storage.MetricValue{
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

func UpdateNotImplementedPost(rw http.ResponseWriter, request *http.Request) {
	log.Println("Update not implemented statType")

	rw.WriteHeader(http.StatusNotImplemented)
	rw.Write([]byte("Not implemented"))
}

func PrintStatsValues(rw http.ResponseWriter, request *http.Request, metricsMemoryRepo storage.MetricStorager, templatesPath string) {
	t, err := template.ParseFiles(templatesPath)
	if err != nil {
		log.Println("Cant parse template ", err)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(rw, metricsMemoryRepo.ReadAll())
	if err != nil {
		log.Println("Cant render template ", err)
		return
	}
}

//JSONStatValue get stat value via json
func JSONStatValue(rw http.ResponseWriter, request *http.Request, metricsMemoryRepo storage.MetricStorager, SignKey string) {
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

	statValue, err := metricsMemoryRepo.Read(InputMetricsJSON.ID, InputMetricsJSON.MType)
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

	if SignKey != "" {
		answerJSON.Hash = hex.EncodeToString(answerJSON.Metric.GetHash(InputMetricsJSON.ID, SignKey))
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	err = json.NewEncoder(rw).Encode(answerJSON)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func PrintStatValue(rw http.ResponseWriter, request *http.Request, metricsMemoryRepo storage.MetricStorager) {
	statType := chi.URLParam(request, "statType")
	statName := chi.URLParam(request, "statName")

	metric, err := metricsMemoryRepo.Read(statName, statType)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Unknown statName"))
		return
	}

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(metric.GetStringValue()))
}
