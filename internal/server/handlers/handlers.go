package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi"
	"metrics/internal/server/storage"
)

type Metric struct {
	ID    string   `json:"id" valid:"required"`
	MType string   `json:"type" valid:"required,in(counter|gauge)"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

//UpdateStatJSONPost update stat via json
func UpdateStatJSONPost(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	var OneMetric Metric

	err := json.NewDecoder(request.Body).Decode(&OneMetric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = govalidator.ValidateStruct(OneMetric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if OneMetric.MType == "counter" {
		if OneMetric.Delta == nil {
			http.Error(rw, "delta is empty", http.StatusBadRequest)
			return
		}

		err = memStatsStorage.UpdateCounterValue(OneMetric.ID, *OneMetric.Delta)
	}

	if OneMetric.MType == "gauge" {
		if OneMetric.Value == nil {
			http.Error(rw, "value is empty", http.StatusBadRequest)
			return
		}

		err = memStatsStorage.UpdateGaugeValue(OneMetric.ID, *OneMetric.Value)
	}

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	log.Println("Update metric via JSON")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Ok"))
}

func UpdateGaugePost(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statValueInt, err := strconv.ParseFloat(statValue, 64)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Bad request"))
		return
	}

	err = memStatsStorage.UpdateGaugeValue(statName, statValueInt)
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

func UpdateCounterPost(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statCounterValue, err := strconv.ParseInt(statValue, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	err = memStatsStorage.UpdateCounterValue(statName, statCounterValue)
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

func PrintStatsValues(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	htmlTemplate := `
<html>
    <head>
    <title></title>
    </head>
    <body>
		<h3 class="keyvalues-header">All values: </h3>
        %v
    </body>
</html>`
	keyValuesHTML := ""

	for k, v := range memStatsStorage.GetDBSchema() {
		keyValuesHTML += fmt.Sprintf("<div><b>%v</b>: %v</div>", k, v)
	}

	htmlPage := fmt.Sprintf(htmlTemplate, keyValuesHTML)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(htmlPage))
}

//JSONStatValue get stat value via json
func JSONStatValue(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
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

	statValue, err := memStatsStorage.ReadValue(InputMetricsJSON.ID)
	if err != nil {
		http.Error(rw, "Unknown statName", http.StatusNotFound)
		return
	}

	answerJSON := Metric{
		ID:    InputMetricsJSON.ID,
		MType: InputMetricsJSON.MType,
	}

	if answerJSON.MType == "counter" {
		var metricValue int64
		metricValue, err = strconv.ParseInt(statValue, 10, 64)
		answerJSON.Delta = &metricValue
	} else {
		var metricValue float64
		metricValue, err = strconv.ParseFloat(statValue, 64)
		answerJSON.Value = &metricValue
	}
	if err != nil {
		http.Error(rw, "Server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	err = json.NewEncoder(rw).Encode(answerJSON)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func PrintStatValue(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	statName := chi.URLParam(request, "statName")
	statValue, err := memStatsStorage.ReadValue(statName)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Unknown statName"))
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(statValue))
}
