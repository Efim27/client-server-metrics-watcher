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

//UpdateStatJSONPost update stat via json
func UpdateStatJSONPost(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	var OneMetric storage.Metric

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

	err = memStatsStorage.Update(OneMetric.ID, &storage.MetricValue{
		MType: OneMetric.MType,
		Value: OneMetric.Value,
		Delta: OneMetric.Delta,
	})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Update metric via JSON")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Ok"))
}

func UpdateGaugePost(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statValueFloat, err := strconv.ParseFloat(statValue, 64)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Bad request"))
		return
	}

	err = memStatsStorage.Update(statName, &storage.MetricValue{
		MType: "gauge",
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

func UpdateCounterPost(rw http.ResponseWriter, request *http.Request, memStatsStorage storage.MemStatsMemoryRepo) {
	statName := chi.URLParam(request, "statName")
	statValue := chi.URLParam(request, "statValue")
	statValueInt, err := strconv.ParseInt(statValue, 10, 64)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	err = memStatsStorage.Update(statName, &storage.MetricValue{
		MType: "counter",
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

	for metricKey, metric := range memStatsStorage.GetAllMetrics() {
		keyValuesHTML += fmt.Sprintf("<div><b>%v</b>: %v</div>", metricKey, metric.GetStringValue())
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

	statValue, err := memStatsStorage.ReadValue(InputMetricsJSON.ID, InputMetricsJSON.MType)
	if err != nil {
		http.Error(rw, "Unknown statName", http.StatusNotFound)
		return
	}

	answerJSON := storage.Metric{
		ID:    InputMetricsJSON.ID,
		MType: InputMetricsJSON.MType,
		Delta: statValue.Delta,
		Value: statValue.Value,
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
	statType := chi.URLParam(request, "statType")
	statName := chi.URLParam(request, "statName")

	metric, err := memStatsStorage.ReadValue(statName, statType)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Unknown statName"))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(metric.GetStringValue()))
}
