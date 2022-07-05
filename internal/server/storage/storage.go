package storage

const (
	MeticTypeGauge   = "gauge"
	MeticTypeCounter = "counter"
)

type MetricMap map[string]MetricValue

type MetricStorager interface {
	UpdateMany(DBSchema map[string]MetricValue)
	InitFromFile()
	Update(key string, value MetricValue) error
	Read(key string, metricType string) (MetricValue, error)
	ReadAll() map[string]MetricMap
	Close() error
	Ping() error
}
