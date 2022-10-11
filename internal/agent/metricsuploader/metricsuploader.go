// Package metricsuploader - HTTP клиент для отправки метрик на сервер.
package metricsuploader

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"metrics/internal/agent/config"
	"metrics/internal/agent/statsreader"
	handlerRSA "metrics/internal/rsa"
	"metrics/internal/server/storage"
)

type MetricsUplader struct {
	client       *resty.Client
	config       config.HTTPClientConfig
	publicKeyRSA *rsa.PublicKey
	protocol     string // http/https
	signKey      string
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

func NewMetricsUploader(config config.HTTPClientConfig, signKey, publicKeyRSA string) *MetricsUplader {
	var metricsUplader MetricsUplader
	metricsUplader.config = config
	metricsUplader.signKey = signKey
	metricsUplader.protocol = "http"
	client := resty.New()

	client.
		SetRetryCount(metricsUplader.config.RetryCount).
		SetRetryWaitTime(metricsUplader.config.RetryWaitTime).
		SetRetryMaxWaitTime(metricsUplader.config.RetryMaxWaitTime)
	metricsUplader.client = client

	if publicKeyRSA != "" {
		var err error
		metricsUplader.publicKeyRSA, err = handlerRSA.ParsePublicKeyRSA(publicKeyRSA)
		if err != nil {
			log.Fatal("Parsing public key failed, RSA disabled")
		}
	}

	if metricsUplader.config.IsEnabledHTTP {
		err := metricsUplader.addCertCA()
		if err != nil {
			log.Println("Error while adding CA: ", err)
			return &metricsUplader
		}

		metricsUplader.protocol = "https"
	}

	return &metricsUplader
}

func (metricsUplader *MetricsUplader) addCertCA() (err error) {
	caCert, err := ioutil.ReadFile("./keysSSL/server.crt")
	if err != nil {
		return
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	metricsUplader.client.SetTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	})

	return
}

// oneStatUploadJSON - отправка 1 метрики.
// Deprecated: используйте MetricsUploadBatch
func (metricsUplader *MetricsUplader) oneStatUpload(statType string, statName string, statValue string) error {
	resp, err := metricsUplader.client.R().
		SetPathParams(map[string]string{
			"addr":     metricsUplader.config.ServerAddr,
			"protocol": metricsUplader.protocol,
			"type":     statType,
			"name":     statName,
			"value":    statValue,
		}).
		SetHeader("Content-Type", "text/plain").
		Post("{protocol}://{addr}/update/{type}/{name}/{value}")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New("HTTP Status != 200")
	}

	return nil
}

// oneStatUploadJSON - отправка 1 метрики в формате JSON.
// Deprecated: используйте MetricsUploadBatch
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
			"addr":     metricsUplader.config.ServerAddr,
			"protocol": metricsUplader.protocol,
		}).
		Post("{protocol}://{addr}/update/")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("HTTP Status: %v (not 200)", resp.StatusCode())
	}

	return nil
}

// MetricsUploadSync - конкурентная отправка метрик.
// Deprecated: используйте MetricsUploadBatch
func (metricsUplader *MetricsUplader) MetricsUploadSync(metricsDump statsreader.MetricsDump) (err error) {
	metricsDump.RLock()
	defer metricsDump.RUnlock()

	for key, metricRawValue := range metricsDump.MetricsGauge {
		metricName := key
		metricValue := fmt.Sprintf("%v", metricRawValue)

		err = metricsUplader.oneStatUploadJSON("gauge", metricName, metricValue)
		if err != nil {
			return
		}
	}

	for key, metricRawValue := range metricsDump.MetricsCounter {
		metricName := key
		metricValue := fmt.Sprintf("%v", metricRawValue)

		err = metricsUplader.oneStatUploadJSON("counter", metricName, metricValue)
		if err != nil {
			return
		}
	}

	return
}

// MetricsUploadAsync - конкурентная отправка метрик.
// Deprecated: используйте MetricsUploadBatch
func (metricsUplader *MetricsUplader) MetricsUploadAsync(metricsDump statsreader.MetricsDump) error {
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

// MetricsUploadBatch - отправка метрик 1 запросом в формате JSON.
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

	if metricsUplader.publicKeyRSA != nil {
		statJSON = handlerRSA.EncryptWithPublicKey(statJSON, metricsUplader.publicKeyRSA)
	}

	metricsUplader.client.SetTransport(&http.Transport{})
	resp, err := metricsUplader.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(statJSON)).
		SetPathParams(map[string]string{
			"addr":     metricsUplader.config.ServerAddr,
			"protocol": metricsUplader.protocol,
		}).
		Post("{protocol}://{addr}/updates/")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("HTTP Status: %v (not 200)", resp.StatusCode())
	}

	return nil
}
