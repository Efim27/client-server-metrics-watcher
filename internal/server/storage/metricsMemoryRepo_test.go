package storage

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"metrics/internal/server/config"
)

func ExampleMemoryRepo() {
	memoryRepo, err := NewMemoryRepo()
	if err != nil {
		log.Fatal(err)
	}

	var counterValueExpect int64 = 50
	err = memoryRepo.Write("PollCount", MetricValue{
		MType: MeticTypeCounter,
		Delta: &counterValueExpect,
	})
	if err != nil {
		log.Fatal(err)
	}

	counterValueReal, err := memoryRepo.Read("PollCount")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(counterValueExpect == *counterValueReal.Delta)
}

func TestMemoryRepoRW(t *testing.T) {
	memoryRepo, err := NewMemoryRepo()
	require.NoError(t, err)

	var counterValueExpect int64 = 50
	err = memoryRepo.Write("PollCount", MetricValue{
		MType: MeticTypeCounter,
		Delta: &counterValueExpect,
	})
	require.NoError(t, err)
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
		MType: MeticTypeCounter,
		Delta: &startValue,
	})
	require.NoError(t, err)
	err = metricsMemoryRepo.Update("PollCount", MetricValue{
		MType: MeticTypeCounter,
		Delta: &incrementValue,
	})
	require.NoError(t, err)
	PollCount, err := metricsMemoryRepo.Read("PollCount", MeticTypeCounter)
	require.NoError(t, err)

	require.Equal(t, int64(29), *PollCount.Delta)
}
