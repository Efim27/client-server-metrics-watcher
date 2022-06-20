package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"metrics/internal/server/config"
)

const syncUploadSymbol = "O"

type MetricValue struct {
	MType string   `json:"type" valid:"required,in(counter|gauge)"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Metric struct {
	ID    string   `json:"id" valid:"required"`
	MType string   `json:"type" valid:"required,in(counter|gauge)"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (metric MetricValue) GetStringValue() string {
	switch metric.MType {
	case "gauge":
		return fmt.Sprintf("%v", *metric.Value)
	case "counter":
		return fmt.Sprintf("%v", *metric.Delta)
	default:
		return ""
	}
}

type MetricStorager interface {
	Len() int
	Write(key string, value *MetricValue) error
	Read(key string) (*MetricValue, error)
	Delete(key string) (*MetricValue, bool)
	GetSchemaDump() map[string]*MetricValue
	Close() error
}

// MemoryRepo структура
type MemoryRepo struct {
	db map[string]*MetricValue
	*sync.RWMutex
}

func NewMemoryRepo() (*MemoryRepo, error) {
	return &MemoryRepo{
		db:      make(map[string]*MetricValue),
		RWMutex: &sync.RWMutex{},
	}, nil
}

func (m *MemoryRepo) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.db)
}

func (m MemoryRepo) Write(key string, value *MetricValue) error {
	m.Lock()
	defer m.Unlock()
	m.db[key] = value
	return nil
}

func (m *MemoryRepo) Delete(key string) (*MetricValue, bool) {
	m.Lock()
	defer m.Unlock()
	oldValue, ok := m.db[key]
	if ok {
		delete(m.db, key)
	}
	return oldValue, ok
}

func (m MemoryRepo) Read(key string) (*MetricValue, error) {
	m.RLock()
	defer m.RUnlock()
	value, err := m.db[key]
	if !err {
		return nil, errors.New("Значение по ключу не найдено, ключ: " + key)
	}

	return value, nil
}

func (m MemoryRepo) GetSchemaDump() map[string]*MetricValue {
	m.RLock()
	defer m.RUnlock()
	return m.db
}

func (m *MemoryRepo) Close() error {
	return nil
}

//MemStatsMemoryRepo - репо для приходящей статистики
type MemStatsMemoryRepo struct {
	uploadMutex    *sync.RWMutex
	gaugeStorage   MetricStorager
	counterStorage MetricStorager
}

func NewMemStatsMemoryRepo() MemStatsMemoryRepo {
	var memStatsStorage MemStatsMemoryRepo
	var err error

	memStatsStorage.uploadMutex = &sync.RWMutex{}
	memStatsStorage.gaugeStorage, err = NewMemoryRepo()
	if err != nil {
		panic("gaugeMemoryRepo init error")
	}
	memStatsStorage.counterStorage, err = NewMemoryRepo()
	if err != nil {
		panic("counterMemoryRepo init error")
	}

	if config.AppConfig.Store.Interval != syncUploadSymbol {
		memStatsStorage.IterativeUploadToFile()
	}

	return memStatsStorage
}

func (memStatsStorage MemStatsMemoryRepo) Update(key string, newMetricValue *MetricValue) error {

	switch newMetricValue.MType {
	case "gauge":
		if newMetricValue.Value == nil {
			return errors.New("MetricValue Value is empty")
		}
		newMetricValue.Delta = nil

		return memStatsStorage.updateGaugeValue(key, newMetricValue)
	case "counter":
		if newMetricValue.Delta == nil {
			return errors.New("MetricValue Delta is empty")
		}
		newMetricValue.Value = nil

		return memStatsStorage.updateCounterValue(key, newMetricValue)
	}

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) updateGaugeValue(key string, newMetricValue *MetricValue) error {
	memStatsStorage.uploadMutex.Lock()
	err := memStatsStorage.gaugeStorage.Write(key, newMetricValue)
	memStatsStorage.uploadMutex.Unlock()

	if err != nil {
		return err
	}

	if config.AppConfig.Store.Interval == syncUploadSymbol {
		return memStatsStorage.UploadToFile()
	}

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) updateCounterValue(key string, newMetricValue *MetricValue) error {
	//Чтение старого значения
	oldMetricValue, err := memStatsStorage.ReadValue(key, "counter")
	if err != nil {
		var delta int64 = 0
		oldMetricValue = &MetricValue{
			Delta: &delta,
		}
	}

	newValue := *oldMetricValue.Delta + *newMetricValue.Delta
	newMetricValue.Delta = &newValue

	memStatsStorage.uploadMutex.Lock()
	memStatsStorage.counterStorage.Write(key, newMetricValue)
	memStatsStorage.uploadMutex.Unlock()

	if config.AppConfig.Store.Interval == syncUploadSymbol {
		return memStatsStorage.UploadToFile()
	}

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) ReadValue(key string, metricType string) (*MetricValue, error) {
	switch metricType {
	case "gauge":
		return memStatsStorage.gaugeStorage.Read(key)
	case "counter":
		return memStatsStorage.counterStorage.Read(key)
	default:
		return nil, errors.New("metricType not found")
	}
}

func (memStatsStorage MemStatsMemoryRepo) UploadToFile() error {
	memStatsStorage.uploadMutex.Lock()
	defer memStatsStorage.uploadMutex.Unlock()
	if config.AppConfig.Store.File == "" {
		return nil
	}

	file, err := os.OpenFile(config.AppConfig.Store.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	allStates := memStatsStorage.GetAllMetrics()
	json.NewEncoder(file).Encode(allStates)

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) IterativeUploadToFile() error {
	intervalSeconds, err := strconv.Atoi(config.AppConfig.Store.Interval)
	if err != nil {
		return err
	}

	tickerUpload := time.NewTicker(time.Duration(intervalSeconds) * time.Second)

	go func() {
		for range tickerUpload.C {
			memStatsStorage.UploadToFile()
		}
	}()

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) InitFromFile() {
	file, err := os.OpenFile(config.AppConfig.Store.File, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	var stateValues map[string]*MetricValue
	json.NewDecoder(file).Decode(&stateValues)

	memStatsStorage.InitStateValues(stateValues)
}

func (memStatsStorage MemStatsMemoryRepo) InitStateValues(DBSchema map[string]*MetricValue) {
	for metricKey, metricValue := range DBSchema {
		memStatsStorage.Update(metricKey, metricValue)
	}
}

func (memStatsStorage MemStatsMemoryRepo) GetAllMetrics() map[string]*MetricValue {
	allMetrics := memStatsStorage.gaugeStorage.GetSchemaDump()

	for metricKey, metricValue := range memStatsStorage.counterStorage.GetSchemaDump() {
		allMetrics[metricKey] = metricValue
	}
	return allMetrics
}

func (memStatsStorage MemStatsMemoryRepo) Close() error {
	err := memStatsStorage.gaugeStorage.Close()
	if err != nil {
		return err
	}
	err = memStatsStorage.counterStorage.Close()

	return err
}
