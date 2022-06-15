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
	m.RLock()
	defer m.RUnlock()
	return m.db
}

func (m *MemoryRepo) Close() error {
	return nil
}

//MemStatsMemoryRepo - репо для приходящей статистики
type MemStatsMemoryRepo struct {
	uploadMutex *sync.RWMutex
	storage     Storager
}

func NewMemStatsMemoryRepo() MemStatsMemoryRepo {
	var memStatsStorage MemStatsMemoryRepo
	var err error

	memStatsStorage.uploadMutex = &sync.RWMutex{}
	memStatsStorage.storage, err = NewMemoryRepo()
	if err != nil {
		panic("MemoryRepo init error")
	}

	if config.AppConfig.Store.Restore {
		memStatsStorage.InitFromFile()
	}

	return memStatsStorage
}

func (memStatsStorage MemStatsMemoryRepo) UpdateGaugeValue(key string, value float64) error {
	memStatsStorage.uploadMutex.Lock()
	err := memStatsStorage.storage.Write(key, fmt.Sprintf("%v", value))
	memStatsStorage.uploadMutex.Unlock()

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
	memStatsStorage.uploadMutex.Lock()
	memStatsStorage.storage.Write(key, newValue)
	memStatsStorage.uploadMutex.Unlock()

	if config.AppConfig.Store.Interval == "O" {
		return memStatsStorage.UploadToFile()
	}

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) ReadValue(key string) (string, error) {
	return memStatsStorage.storage.Read(key)
}

func (memStatsStorage MemStatsMemoryRepo) UploadToFile() error {
	memStatsStorage.uploadMutex.Lock()
	defer memStatsStorage.uploadMutex.Unlock()
	if config.AppConfig.Store.File == "" {
		return nil
	}

	file, err := os.OpenFile(config.AppConfig.Store.File, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	defer file.Close()
	if err != nil {
		return err
	}
	allStates := memStatsStorage.GetDBSchema()
	//log.Println(allStates)
	json.NewEncoder(file).Encode(allStates)

	return nil
}

func (memStatsStorage MemStatsMemoryRepo) InitFromFile() {
	file, err := os.OpenFile(config.AppConfig.Store.File, os.O_RDONLY|os.O_CREATE, 0777)
	defer file.Close()
	if err != nil {
		panic("Error while restoring StateValues")
	}

	var stateValues map[string]string
	json.NewDecoder(file).Decode(&stateValues)
	memStatsStorage.InitStateValues(stateValues)
}

func (memStatsStorage MemStatsMemoryRepo) InitStateValues(DBSchema map[string]string) {
	for stateKey, stateValue := range DBSchema {
		memStatsStorage.storage.Write(stateKey, stateValue)
	}
}

func (memStatsStorage MemStatsMemoryRepo) GetDBSchema() map[string]string {
	return memStatsStorage.storage.GetSchemaDump()
}

func (memStatsStorage MemStatsMemoryRepo) Close() error {
	return memStatsStorage.storage.Close()
}
