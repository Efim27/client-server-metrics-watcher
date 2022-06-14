package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"metrics/internal/server/config"
)

type Gauge float64
type Counter int64

type Storager interface {
	Len() int
	Write(key, value string) error
	Read(key string) (string, error)
	Delete(key string) (string, bool)
	GetSchemaDump() map[string]string
	Close() error
}

// MemoryRepo структура
type MemoryRepo struct {
	db map[string]string
	*sync.RWMutex
}

func NewMemoryRepo() (*MemoryRepo, error) {
	file, err := os.OpenFile(config.AppConfig.Store.File, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return nil, err
	}

	repoEncoderJSON := json.NewEncoder(file)
	repoEncoderJSON.SetIndent("", "  ")

	return &MemoryRepo{
		db:      make(map[string]string),
		RWMutex: &sync.RWMutex{},
	}, nil
}

func (m *MemoryRepo) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.db)
}

func (m MemoryRepo) Write(key, value string) error {
	m.Lock()
	defer m.Unlock()
	m.db[key] = value
	return nil
}

func (m *MemoryRepo) Delete(key string) (string, bool) {
	m.Lock()
	defer m.Unlock()
	oldValue, ok := m.db[key]
	if ok {
		delete(m.db, key)
	}
	return oldValue, ok
}

func (m MemoryRepo) Read(key string) (string, error) {
	m.RLock()
	defer m.RUnlock()
	value, err := m.db[key]
	if !err {
		return "", errors.New("Значение по ключу не найдено, ключ: " + key)
	}

	return value, nil
}

func (m MemoryRepo) GetSchemaDump() map[string]string {
	return m.db
}

func (m *MemoryRepo) Close() error {
	return nil
}

//MemStatsMemoryRepo - репо для приходящей статистики
type MemStatsMemoryRepo struct {
	storage Storager
}

func NewMemStatsMemoryRepo() MemStatsMemoryRepo {
	var memStatsStorage MemStatsMemoryRepo
	var err error
	memStatsStorage.storage, err = NewMemoryRepo()
	if err != nil {
		panic("MemoryRepo init error")
	}

	if config.AppConfig.Store.Restore {
		memStatsStorage.LoadFromFile()
	}

	return memStatsStorage
}

func (memStatsStorage MemStatsMemoryRepo) UpdateGaugeValue(key string, value float64) error {
	err := memStatsStorage.storage.Write(key, fmt.Sprintf("%v", value))
	if err != nil {
		return err
	}

	if config.AppConfig.Store.Interval == "O" {
		return memStatsStorage.UploadToFile()
	}

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) UpdateCounterValue(key string, value int64) error {
	//Чтение старого значения
	oldValue, err := memStatsStorage.storage.Read(key)
	if err != nil {
		oldValue = "0"
	}

	//Конвертация в число
	oldValueInt, err := strconv.ParseInt(oldValue, 10, 64)
	if err != nil {
		return errors.New("MemStats value is not int64")
	}

	newValue := fmt.Sprintf("%v", oldValueInt+value)
	memStatsStorage.storage.Write(key, newValue)

	if config.AppConfig.Store.Interval == "O" {
		return memStatsStorage.UploadToFile()
	}

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) UploadToFile() error {
	if config.AppConfig.Store.File == "" {
		return nil
	}

	file, err := os.OpenFile(config.AppConfig.Store.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	defer file.Close()
	if err != nil {
		return err
	}
	allStates := memStatsStorage.GetDBSchema()
	json.NewEncoder(file).Encode(allStates)

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) LoadFromFile() error {
	return nil
}

func (memStatsStorage MemStatsMemoryRepo) ReadValue(key string) (string, error) {
	return memStatsStorage.storage.Read(key)
}

func (memStatsStorage MemStatsMemoryRepo) GetDBSchema() map[string]string {
	return memStatsStorage.storage.GetSchemaDump()
}

func (memStatsStorage MemStatsMemoryRepo) Close() error {
	return memStatsStorage.storage.Close()
}
