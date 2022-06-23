package storage

type MetricStorager interface {
	Update(key string, value MetricValue) error
	Read(key string, metricType string) (MetricValue, error)
	ReadAll() map[string]MetricValue
	Close() error
}
