package storage

import (
	"errors"
	"sync"
)

type MemoryRepo struct {
	db map[string]MetricValue
	*sync.RWMutex
}

func NewMemoryRepo() (*MemoryRepo, error) {
	return &MemoryRepo{
		db:      make(map[string]MetricValue),
		RWMutex: &sync.RWMutex{},
	}, nil
}

func (m *MemoryRepo) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.db)
}

func (m MemoryRepo) Write(key string, value MetricValue) error {
	m.Lock()
	defer m.Unlock()
	m.db[key] = value
	return nil
}

func (m *MemoryRepo) Delete(key string) (MetricValue, bool) {
	m.Lock()
	defer m.Unlock()
	oldValue, ok := m.db[key]
	if ok {
		delete(m.db, key)
	}
	return oldValue, ok
}

func (m MemoryRepo) Read(key string) (MetricValue, error) {
	m.RLock()
	defer m.RUnlock()
	value, err := m.db[key]
	if !err {
		return MetricValue{}, errors.New("Значение по ключу не найдено, ключ: " + key)
	}

	return value, nil
}

func (m MemoryRepo) GetSchemaDump() map[string]MetricValue {
	m.RLock()
	defer m.RUnlock()
	return m.db
}

func (m *MemoryRepo) Close() error {
	return nil
}
