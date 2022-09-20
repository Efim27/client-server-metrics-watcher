// Package statsreader - считыватель runtime метрик
package statsreader

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type gauge float64
type counter int64

// MetricsDump - потокобезопасное хранилище метрик.
type MetricsDump struct {
	*sync.RWMutex
	MetricsGauge   map[string]gauge
	MetricsCounter map[string]counter
}

func NewMetricsDump() (*MetricsDump, error) {
	return &MetricsDump{
		RWMutex:        &sync.RWMutex{},
		MetricsGauge:   make(map[string]gauge),
		MetricsCounter: make(map[string]counter),
	}, nil
}

// Refresh - считыватель метрик.
func (metricsDump *MetricsDump) Refresh() {
	var MemStatistics runtime.MemStats
	runtime.ReadMemStats(&MemStatistics)

	metricsDump.Lock()
	defer metricsDump.Unlock()

	metricsDump.MetricsGauge["BuckHashSys"] = gauge(MemStatistics.BuckHashSys)
	metricsDump.MetricsGauge["Frees"] = gauge(MemStatistics.Frees)
	metricsDump.MetricsGauge["GCCPUFraction"] = gauge(MemStatistics.GCCPUFraction)
	metricsDump.MetricsGauge["GCSys"] = gauge(MemStatistics.GCSys)
	metricsDump.MetricsGauge["HeapAlloc"] = gauge(MemStatistics.HeapAlloc)

	metricsDump.MetricsGauge["HeapIdle"] = gauge(MemStatistics.HeapIdle)
	metricsDump.MetricsGauge["HeapInuse"] = gauge(MemStatistics.HeapInuse)
	metricsDump.MetricsGauge["HeapObjects"] = gauge(MemStatistics.HeapObjects)
	metricsDump.MetricsGauge["HeapReleased"] = gauge(MemStatistics.HeapReleased)
	metricsDump.MetricsGauge["HeapSys"] = gauge(MemStatistics.HeapSys)

	metricsDump.MetricsGauge["LastGC"] = gauge(MemStatistics.LastGC)
	metricsDump.MetricsGauge["Lookups"] = gauge(MemStatistics.Lookups)
	metricsDump.MetricsGauge["MCacheInuse"] = gauge(MemStatistics.MCacheInuse)
	metricsDump.MetricsGauge["MCacheSys"] = gauge(MemStatistics.MCacheSys)
	metricsDump.MetricsGauge["MSpanInuse"] = gauge(MemStatistics.MSpanInuse)

	metricsDump.MetricsGauge["MSpanSys"] = gauge(MemStatistics.MSpanSys)
	metricsDump.MetricsGauge["Mallocs"] = gauge(MemStatistics.Mallocs)
	metricsDump.MetricsGauge["NextGC"] = gauge(MemStatistics.NextGC)
	metricsDump.MetricsGauge["NumForcedGC"] = gauge(MemStatistics.NumForcedGC)
	metricsDump.MetricsGauge["NumGC"] = gauge(MemStatistics.NumGC)

	metricsDump.MetricsGauge["OtherSys"] = gauge(MemStatistics.OtherSys)
	metricsDump.MetricsGauge["PauseTotalNs"] = gauge(MemStatistics.PauseTotalNs)
	metricsDump.MetricsGauge["StackInuse"] = gauge(MemStatistics.StackInuse)
	metricsDump.MetricsGauge["StackSys"] = gauge(MemStatistics.StackSys)

	metricsDump.MetricsGauge["Alloc"] = gauge(MemStatistics.Alloc)
	metricsDump.MetricsGauge["Sys"] = gauge(MemStatistics.Sys)
	metricsDump.MetricsGauge["TotalAlloc"] = gauge(MemStatistics.TotalAlloc)
	metricsDump.MetricsGauge["RandomValue"] = gauge(rand.Float64())

	metricsDump.MetricsCounter["PollCount"] = metricsDump.MetricsCounter["PollCount"] + 1
}

// RefreshExtra - считыватель дополнительных метрик.
func (metricsDump *MetricsDump) RefreshExtra() error {
	metrics, err := mem.VirtualMemory()
	if err != nil {
		return nil
	}

	metricsDump.Lock()
	defer metricsDump.Unlock()

	metricsDump.MetricsGauge["TotalMemory"] = gauge(metrics.Total)
	metricsDump.MetricsGauge["FreeMemory"] = gauge(metrics.Free)

	percentageCPU, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	for i, currentPercentageCPU := range percentageCPU {
		metricName := fmt.Sprintf("CPUutilization%v", i)
		metricsDump.MetricsGauge[metricName] = gauge(currentPercentageCPU)
	}

	return nil
}
