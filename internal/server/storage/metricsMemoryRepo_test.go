package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"metrics/internal/server/config"
)

func TestMemoryRepoRW(t *testing.T) {
	memoryRepo, err := NewMemoryRepo()
	require.NoError(t, err)

	var counterValueExpect int64 = 50
	memoryRepo.Write("PollCount", MetricValue{
		MType: "counter",
		Delta: &counterValueExpect,
	})
	counterValueReal, err := memoryRepo.Read("PollCount")

	require.NoError(t, err)
	require.Equal(t, counterValueExpect, *counterValueReal.Delta)
}

func TestMemoryRepoReadEmpty(t *testing.T) {
	memoryRepo, err := NewMemoryRepo()
	require.NoError(t, err)
	_, err = memoryRepo.Read("username")
	require.Error(t, err)
}

func TestUpdateCounterValue(t *testing.T) {
	metricsMemoryRepo := NewMetricsMemoryRepo(config.StoreConfig{})

	var startValue int64 = 7
	var incrementValue int64 = 22
	err := metricsMemoryRepo.Update("PollCount", MetricValue{
		MType: "counter",
		Delta: &startValue,
	})
	require.NoError(t, err)
	err = metricsMemoryRepo.Update("PollCount", MetricValue{
		MType: "counter",
		Delta: &incrementValue,
	})
	require.NoError(t, err)
	PollCount, err := metricsMemoryRepo.Read("PollCount", "counter")
	require.NoError(t, err)

	require.Equal(t, int64(29), *PollCount.Delta)
}
