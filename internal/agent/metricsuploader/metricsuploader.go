package metricsuploader

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"metrics/internal/agent/config"
	"metrics/internal/agent/statsreader"
)

type MetricsUplader struct {
	client *resty.Client
	config config.HTTPClientConfig
}

func NewMetricsUploader(config config.HTTPClientConfig) *MetricsUplader {
	var metricsUplader MetricsUplader
	metricsUplader.config = config
	client := resty.New()

	client.
		SetRetryCount(metricsUplader.config.RetryCount).
		SetRetryWaitTime(metricsUplader.config.RetryWaitTime).
		SetRetryMaxWaitTime(metricsUplader.config.RetryMaxWaitTime)
	metricsUplader.client = client

	return &metricsUplader
}

func (metricsUplader *MetricsUplader) oneStatUpload(statType string, statName string, statValue string) error {
	resp, err := metricsUplader.client.R().
		SetPathParams(map[string]string{
			"addr":  metricsUplader.config.ServerAddr,
			"type":  statType,
			"name":  statName,
			"value": statValue,
		}).
		SetHeader("Content-Type", "text/plain").
		Post("http://{addr}/update/{type}/{name}/{value}")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New("HTTP Status != 200")
	}

	return nil
}

func (metricsUplader *MetricsUplader) oneStatUploadJSON(statType string, statName string, statValue string) error {
	OneMetrics := struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Delta int64   `json:"delta"`
		Value float64 `json:"value"`
	}{
		ID:    statName,
		MType: statType,
	}

	var err error
	switch OneMetrics.MType {
	case "counter":
		var metricValue int64
		metricValue, err = strconv.ParseInt(statValue, 10, 64)
		OneMetrics.Delta = metricValue
	case "gauge":
		var metricValue float64
		metricValue, err = strconv.ParseFloat(statValue, 64)
		OneMetrics.Value = metricValue
	default:
		return errors.New("unknown statType")
	}
	if err != nil {
		return errors.New("invalid statValue")
	}

	statJSON, err := json.Marshal(OneMetrics)
	if err != nil {
		return err
	}

	resp, err := metricsUplader.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(statJSON)).
		SetPathParams(map[string]string{
			"addr": metricsUplader.config.ServerAddr,
		}).
		Post("http://{addr}/update/")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("HTTP Status: %v (not 200)", resp.StatusCode())
	}

	return nil
}

func (metricsUplader *MetricsUplader) MemoryStatsUpload(memoryStats statsreader.MemoryStatsDump) error {
	reflectMemoryStats := reflect.ValueOf(memoryStats)
	typeOfMemoryStats := reflectMemoryStats.Type()
	errorGroup := new(errgroup.Group)

	for i := 0; i < reflectMemoryStats.NumField(); i++ {
		statName := typeOfMemoryStats.Field(i).Name
		statValue := fmt.Sprintf("%v", reflectMemoryStats.Field(i).Interface())
		statType := strings.Split(typeOfMemoryStats.Field(i).Type.String(), ".")[1]

		errorGroup.Go(func() error {
			return metricsUplader.oneStatUploadJSON(statType, statName, statValue)
		})
	}

	err := errorGroup.Wait()
	return err
}
