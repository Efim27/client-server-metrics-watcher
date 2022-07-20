package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"metrics/internal/server/config"
)

const syncUploadSymbol = time.Duration(0)

type MetricValue struct {
	MType string   `json:"type" valid:"required,in(counter|gauge)"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Metric struct {
	ID string `json:"id" valid:"required"`
	MetricValue
}

func (metric MetricValue) GetStringValue() string {
	switch metric.MType {
	case MeticTypeGauge:
		return fmt.Sprintf("%v", *metric.Value)
	case MeticTypeCounter:
		return fmt.Sprintf("%v", *metric.Delta)
	default:
		return ""
	}
}

func (metric MetricValue) GetHash(id, signKey string) []byte {
	if signKey == "" {
		return nil
	}

	var metricLabel string
	switch metric.MType {
	case MeticTypeGauge:
		metricLabel = fmt.Sprintf("%s:gauge:%f", id, *metric.Value)
	case MeticTypeCounter:
		metricLabel = fmt.Sprintf("%s:counter:%d", id, *metric.Delta)
	default:
		return nil
	}

	signerHMAC := hmac.New(sha256.New, []byte(signKey))
	signerHMAC.Write([]byte(metricLabel))
	return signerHMAC.Sum(nil)
}

//MetricsMemoryRepo - репо в оперативной памяти для приходящей статистики
type MetricsMemoryRepo struct {
	uploadMutex    *sync.RWMutex
	gaugeStorage   *MemoryRepo
	counterStorage *MemoryRepo
	config         config.StoreConfig
}

func NewMetricsMemoryRepo(config config.StoreConfig) MetricsMemoryRepo {
	var metricsMemoryRepo MetricsMemoryRepo
	var err error

	metricsMemoryRepo.config = config
	metricsMemoryRepo.uploadMutex = &sync.RWMutex{}
	metricsMemoryRepo.gaugeStorage, err = NewMemoryRepo()
	if err != nil {
		panic("gaugeMemoryRepo init error")
	}
	metricsMemoryRepo.counterStorage, err = NewMemoryRepo()
	if err != nil {
		panic("counterMemoryRepo init error")
	}

	if metricsMemoryRepo.config.Interval != syncUploadSymbol {
		metricsMemoryRepo.IterativeUploadToFile()
	}

	return metricsMemoryRepo
}

func (metricsMemoryRepo MetricsMemoryRepo) Update(key string, newMetricValue MetricValue) error {
	switch newMetricValue.MType {
	case MeticTypeGauge:
		if newMetricValue.Value == nil {
			return errors.New("metric Value is empty")
		}
		newMetricValue.Delta = nil

		return metricsMemoryRepo.updateGaugeValue(key, newMetricValue)
	case MeticTypeCounter:
		if newMetricValue.Delta == nil {
			return errors.New("metric Delta is empty")
		}
		newMetricValue.Value = nil

		return metricsMemoryRepo.updateCounterValue(key, newMetricValue)
	default:
		return errors.New("metric type is not defined")
	}
}

func (metricsMemoryRepo MetricsMemoryRepo) updateGaugeValue(key string, newMetricValue MetricValue) error {
	metricsMemoryRepo.uploadMutex.Lock()
	err := metricsMemoryRepo.gaugeStorage.Write(key, newMetricValue)
	metricsMemoryRepo.uploadMutex.Unlock()

	if err != nil {
		return err
	}

	if metricsMemoryRepo.config.Interval == syncUploadSymbol {
		return metricsMemoryRepo.UploadToFile()
	}

	return nil
}

func (metricsMemoryRepo MetricsMemoryRepo) updateCounterValue(key string, newMetricValue MetricValue) error {
	//Чтение старого значения
	oldMetricValue, err := metricsMemoryRepo.Read(key, MeticTypeCounter)
	if err != nil {
		var delta int64 = 0
		oldMetricValue = MetricValue{
			Delta: &delta,
		}
	}

	newValue := *oldMetricValue.Delta + *newMetricValue.Delta
	newMetricValue.Delta = &newValue

	metricsMemoryRepo.uploadMutex.Lock()
	metricsMemoryRepo.counterStorage.Write(key, newMetricValue)
	metricsMemoryRepo.uploadMutex.Unlock()

	if metricsMemoryRepo.config.Interval == syncUploadSymbol {
		return metricsMemoryRepo.UploadToFile()
	}

	return nil
}

func (metricsMemoryRepo MetricsMemoryRepo) Read(key string, metricType string) (MetricValue, error) {
	switch metricType {
	case MeticTypeGauge:
		return metricsMemoryRepo.gaugeStorage.Read(key)
	case MeticTypeCounter:
		return metricsMemoryRepo.counterStorage.Read(key)
	default:
		return MetricValue{}, errors.New("metricType not found")
	}
}

func (metricsMemoryRepo MetricsMemoryRepo) UploadToFile() error {
	metricsMemoryRepo.uploadMutex.Lock()
	defer metricsMemoryRepo.uploadMutex.Unlock()
	if metricsMemoryRepo.config.File == "" {
		return nil
	}

	file, err := os.OpenFile(metricsMemoryRepo.config.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	allStates := metricsMemoryRepo.ReadAll()
	json.NewEncoder(file).Encode(allStates)

	return nil
}

func (metricsMemoryRepo MetricsMemoryRepo) IterativeUploadToFile() {
	tickerUpload := time.NewTicker(metricsMemoryRepo.config.Interval)

	go func() {
		for range tickerUpload.C {
			err := metricsMemoryRepo.UploadToFile()
			if err != nil {
				panic(err)
			}
		}
	}()
}

func (metricsMemoryRepo MetricsMemoryRepo) InitFromFile() {
	file, err := os.OpenFile(metricsMemoryRepo.config.File, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	var metricsDump map[string]MetricMap
	err = json.NewDecoder(file).Decode(&metricsDump)
	if err != nil {
		panic(err)
	}

	for _, metricList := range metricsDump {
		err = metricsMemoryRepo.UpdateMany(metricList)
	}
	if err != nil {
		panic(err)
	}
}

func (metricsMemoryRepo MetricsMemoryRepo) UpdateManySliceMetric(MetricBatch []Metric) error {
	for _, metricValue := range MetricBatch {
		err := metricsMemoryRepo.Update(metricValue.ID, metricValue.MetricValue)

		if err != nil {
			return err
		}
	}

	return nil
}

func (metricsMemoryRepo MetricsMemoryRepo) UpdateMany(DBSchema map[string]MetricValue) error {
	for metricKey, metricValue := range DBSchema {
		err := metricsMemoryRepo.Update(metricKey, metricValue)

		if err != nil {
			return err
		}
	}

	return nil
}

func (metricsMemoryRepo MetricsMemoryRepo) ReadAll() map[string]MetricMap {
	return map[string]MetricMap{
		MeticTypeGauge:   metricsMemoryRepo.gaugeStorage.GetSchemaDump(),
		MeticTypeCounter: metricsMemoryRepo.counterStorage.GetSchemaDump(),
	}
}

func (metricsMemoryRepo MetricsMemoryRepo) Close() error {
	err := metricsMemoryRepo.gaugeStorage.Close()
	if err != nil {
		return err
	}
	err = metricsMemoryRepo.counterStorage.Close()

	return err
}

func (metricsMemoryRepo MetricsMemoryRepo) Ping() error {
	if metricsMemoryRepo.gaugeStorage.Ping() != nil {
		return metricsMemoryRepo.gaugeStorage.Ping()
	}

	if metricsMemoryRepo.counterStorage.Ping() != nil {
		return metricsMemoryRepo.gaugeStorage.Ping()
	}

	return nil
}
