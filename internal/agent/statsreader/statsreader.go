package statsreader

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type gauge float64
type counter int64

type MetricsDump struct {
	mu sync.RWMutex

	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge

	HeapAlloc    gauge
	HeapIdle     gauge
	HeapInuse    gauge
	HeapObjects  gauge
	HeapReleased gauge

	HeapSys     gauge
	LastGC      gauge
	Lookups     gauge
	MCacheInuse gauge
	MCacheSys   gauge

	MSpanInuse  gauge
	MSpanSys    gauge
	Mallocs     gauge
	NextGC      gauge
	NumForcedGC gauge

	NumGC        gauge
	OtherSys     gauge
	PauseTotalNs gauge
	StackInuse   gauge
	StackSys     gauge

	Sys         gauge
	TotalAlloc  gauge
	PollCount   counter
	RandomValue gauge

	//Extra
	TotalMemory     gauge
	FreeMemory      gauge
	CPUUtilizations []gauge
}

func (metricsDump *MetricsDump) Refresh() {
	var MemStatistics runtime.MemStats
	runtime.ReadMemStats(&MemStatistics)

	metricsDump.mu.Lock()
	defer metricsDump.mu.Unlock()

	metricsDump.BuckHashSys = gauge(MemStatistics.BuckHashSys)
	metricsDump.Frees = gauge(MemStatistics.Frees)
	metricsDump.GCCPUFraction = gauge(MemStatistics.GCCPUFraction)
	metricsDump.GCSys = gauge(MemStatistics.GCSys)
	metricsDump.HeapAlloc = gauge(MemStatistics.HeapAlloc)

	metricsDump.HeapIdle = gauge(MemStatistics.HeapIdle)
	metricsDump.HeapInuse = gauge(MemStatistics.HeapInuse)
	metricsDump.HeapObjects = gauge(MemStatistics.HeapObjects)
	metricsDump.HeapReleased = gauge(MemStatistics.HeapReleased)
	metricsDump.HeapSys = gauge(MemStatistics.HeapSys)

	metricsDump.LastGC = gauge(MemStatistics.LastGC)
	metricsDump.Lookups = gauge(MemStatistics.Lookups)
	metricsDump.MCacheInuse = gauge(MemStatistics.MCacheInuse)
	metricsDump.MCacheSys = gauge(MemStatistics.MCacheSys)
	metricsDump.MSpanInuse = gauge(MemStatistics.MSpanInuse)

	metricsDump.MSpanSys = gauge(MemStatistics.MSpanSys)
	metricsDump.Mallocs = gauge(MemStatistics.Mallocs)
	metricsDump.NextGC = gauge(MemStatistics.NextGC)
	metricsDump.NumForcedGC = gauge(MemStatistics.NumForcedGC)
	metricsDump.NumGC = gauge(MemStatistics.NumGC)

	metricsDump.OtherSys = gauge(MemStatistics.OtherSys)
	metricsDump.PauseTotalNs = gauge(MemStatistics.PauseTotalNs)
	metricsDump.StackInuse = gauge(MemStatistics.StackInuse)
	metricsDump.StackSys = gauge(MemStatistics.StackSys)

	metricsDump.Alloc = gauge(MemStatistics.Alloc)
	metricsDump.Sys = gauge(MemStatistics.Sys)
	metricsDump.TotalAlloc = gauge(MemStatistics.TotalAlloc)
	metricsDump.PollCount = metricsDump.PollCount + 1
	metricsDump.RandomValue = gauge(rand.Float64())

	metricsDump.RandomValue = gauge(rand.Float64())
}

func (metricsDump *MetricsDump) RefreshExtra() error {
	metrics, err := mem.VirtualMemory()
	if err != nil {
		return nil
	}

	metricsDump.mu.Lock()
	defer metricsDump.mu.Unlock()

	metricsDump.TotalMemory = gauge(metrics.Total)
	metricsDump.FreeMemory = gauge(metrics.Free)

	percentageCPU, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	metricsDump.CPUUtilizations = make([]gauge, len(percentageCPU))
	for i, currentPercentageCPU := range percentageCPU {
		metricsDump.CPUUtilizations[i] = gauge(currentPercentageCPU)
	}

	return nil
}
