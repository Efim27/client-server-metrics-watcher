// Package metricsuploader - HTTP клиент для отправки метрик на сервер
package metricsuploader

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"metrics/internal/agent/config"
	"metrics/internal/agent/statsreader"
	"metrics/internal/server/storage"
)

type MetricsUplader struct {
	client  *resty.Client
	config  config.HTTPClientConfig
	signKey string
}

func newMetricValue(mtype string, value string) (storage.MetricValue, error) {
	mValue := storage.MetricValue{
		MType: mtype,
	}

	var err error
	switch mtype {
	case storage.MeticTypeCounter:
		var metricValue int64
		metricValue, err = strconv.ParseInt(value, 10, 64)
		mValue.Delta = &metricValue
	case storage.MeticTypeGauge:
		var metricValue float64
		metricValue, err = strconv.ParseFloat(value, 64)
		mValue.Value = &metricValue
	default:
		return mValue, errors.New("unknown statType")
	}
	if err != nil {
		return mValue, errors.New("invalid statValue")
	}

	return mValue, nil
}

func NewMetricsUploader(config config.HTTPClientConfig, signKey string) *MetricsUplader {
	var metricsUplader MetricsUplader
	metricsUplader.config = config
	metricsUplader.signKey = signKey
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

func (metricsUplader *MetricsUplader) oneStatUploadJSON(mType string, name string, value string) error {
	metricValue, err := newMetricValue(mType, value)
	if err != nil {
		return nil
	}

	OneMetrics := struct {
		storage.Metric
		Hash string `json:"hash"`
	}{
		Metric: storage.Metric{
			ID:          name,
			MetricValue: metricValue,
		},
	}

	if metricsUplader.signKey != "" {
		OneMetrics.Hash = hex.EncodeToString(OneMetrics.GetHash(OneMetrics.ID, metricsUplader.signKey))
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

// MetricsUpload - конкурентная отправка метрик
// Deprecated: используйте MetricsUploadBatch
func (metricsUplader *MetricsUplader) MetricsUpload(metricsDump statsreader.MetricsDump) error {
	metricsDump.RLock()
	defer metricsDump.RUnlock()

	errorGroup := new(errgroup.Group)

	for key, metricRawValue := range metricsDump.MetricsGauge {
		metricName := key
		metricValue := fmt.Sprintf("%v", metricRawValue)

		errorGroup.Go(func() error {
			return metricsUplader.oneStatUploadJSON("gauge", metricName, metricValue)
		})
	}

	for key, metricRawValue := range metricsDump.MetricsCounter {
		metricName := key
		metricValue := fmt.Sprintf("%v", metricRawValue)

		errorGroup.Go(func() error {
			return metricsUplader.oneStatUploadJSON("counter", metricName, metricValue)
		})
	}

	err := errorGroup.Wait()
	return err
}

// MetricsUploadBatch - отправка метрик 1 запросом в формате JSON
func (metricsUplader *MetricsUplader) MetricsUploadBatch(metricsDump statsreader.MetricsDump) error {
	metricsDump.RLock()
	defer metricsDump.RUnlock()
	var MetricValueBatch []storage.Metric

	for metricName, metricRawValue := range metricsDump.MetricsGauge {
		metricValue := fmt.Sprintf("%v", metricRawValue)

		mValue, err := newMetricValue("gauge", metricValue)
		if err != nil {
			return err
		}

		MetricValueBatch = append(MetricValueBatch, storage.Metric{
			ID:          metricName,
			MetricValue: mValue,
		})
	}

	for metricName, metricRawValue := range metricsDump.MetricsCounter {
		metricValue := fmt.Sprintf("%v", metricRawValue)

		mValue, err := newMetricValue("counter", metricValue)
		if err != nil {
			return err
		}

		MetricValueBatch = append(MetricValueBatch, storage.Metric{
			ID:          metricName,
			MetricValue: mValue,
		})
	}

	statJSON, err := json.Marshal(MetricValueBatch)
	if err != nil {
		return err
	}

	resp, err := metricsUplader.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(statJSON)).
		SetPathParams(map[string]string{
			"addr": metricsUplader.config.ServerAddr,
		}).
		Post("http://{addr}/updates/")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("HTTP Status: %v (not 200)", resp.StatusCode())
	}

	return nil
}
