package storage

import (
	"encoding/hex"
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

	err = memoryRepo.Ping()

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

	err = memoryRepo.Close()
}

func TestMemoryRepoRWD(t *testing.T) {
	memoryRepo, err := NewMemoryRepo()
	require.NoError(t, err)

	err = memoryRepo.Ping()
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

	require.Equal(t, 1, memoryRepo.Len())

	_, ok := memoryRepo.Delete("PollCount")
	require.True(t, ok)

	require.Equal(t, 0, memoryRepo.Len())

	err = memoryRepo.Close()
	require.NoError(t, err)
}

func TestMemoryRepoReadEmpty(t *testing.T) {
	memoryRepo, err := NewMemoryRepo()
	require.NoError(t, err)

	err = memoryRepo.Ping()
	require.NoError(t, err)

	_, err = memoryRepo.Read("username")
	require.Error(t, err)

	require.Equal(t, 0, memoryRepo.Len())

	err = memoryRepo.Close()
	require.NoError(t, err)
}

func TestMemoryRepoUpdateValues(t *testing.T) {
	metricsMemoryRepo := NewMetricsMemoryRepo(config.StoreConfig{})

	err := metricsMemoryRepo.Ping()
	require.NoError(t, err)

	var metricValue1 int64 = 7
	var metricValue2 int64 = 22
	var metricValue3 float64 = 27.5

	err = metricsMemoryRepo.Update("PollCount", MetricValue{
		MType: MeticTypeCounter,
		Delta: &metricValue1,
	})
	require.NoError(t, err)

	err = metricsMemoryRepo.Update("PollCount", MetricValue{
		MType: MeticTypeCounter,
		Delta: &metricValue2,
	})
	require.NoError(t, err)

	err = metricsMemoryRepo.Update("Gauge1", MetricValue{
		MType: MeticTypeGauge,
		Value: &metricValue3,
	})
	require.NoError(t, err)

	PollCount, err := metricsMemoryRepo.Read("PollCount", MeticTypeCounter)
	require.NoError(t, err)
	require.EqualValues(t, 29, *PollCount.Delta)

	Gauge1, err := metricsMemoryRepo.Read("Gauge1", MeticTypeGauge)
	require.NoError(t, err)
	require.EqualValues(t, 27.5, *Gauge1.Value)

	err = metricsMemoryRepo.Close()
	require.NoError(t, err)
}

func TestMemoryRepoReadAll(t *testing.T) {
	metricsMemoryRepo := NewMetricsMemoryRepo(config.StoreConfig{})

	err := metricsMemoryRepo.Ping()
	require.NoError(t, err)

	var metricValueDelta1 int64 = 11
	metricValue1 := MetricValue{
		MType: MeticTypeCounter,
		Delta: &metricValueDelta1,
	}
	err = metricsMemoryRepo.Update("PollCount1", metricValue1)
	require.NoError(t, err)

	var metricValueDelta2 int64 = 22
	metricValue2 := MetricValue{
		MType: MeticTypeCounter,
		Delta: &metricValueDelta2,
	}
	err = metricsMemoryRepo.Update("PollCount2", metricValue2)
	require.NoError(t, err)

	repoValues := metricsMemoryRepo.ReadAll()

	repoMetricMap := MetricMap{"PollCount1": metricValue1, "PollCount2": metricValue2}
	repoValuesExpected := map[string]MetricMap{
		MeticTypeGauge:   {},
		MeticTypeCounter: repoMetricMap,
	}
	require.EqualValues(t, repoValues, repoValuesExpected)

	actualMetricValue1, err := metricsMemoryRepo.Read("PollCount1", MeticTypeCounter)
	require.NoError(t, err)

	actualMetricValue2, err := metricsMemoryRepo.Read("PollCount2", MeticTypeCounter)
	require.NoError(t, err)

	require.Equal(t, "11", actualMetricValue1.GetStringValue())
	require.Equal(t, "22", actualMetricValue2.GetStringValue())

	signKey := "bhHN02mqZa8"
	expectedHash := "f18d2a9843fe5be455fe09a9035b8bfbf7b0dfae5f393c14330f29b804a89e7f"
	actualHash := hex.EncodeToString(actualMetricValue1.GetHash("PollCount1", signKey))
	require.Equal(t, expectedHash, actualHash)

	err = metricsMemoryRepo.Close()
	require.NoError(t, err)
}

func TestMemoryRepoFileIterativeWrite(t *testing.T) {
	metricsMemoryRepo := NewMetricsMemoryRepo(config.StoreConfig{})

	err := metricsMemoryRepo.Ping()
	require.NoError(t, err)

	metricsMemoryRepo.IterativeUploadToFile()

	err = metricsMemoryRepo.Close()
	require.NoError(t, err)
}

func TestMemoryRepoUpdateMany(t *testing.T) {
	metricsMemoryRepo := NewMetricsMemoryRepo(config.StoreConfig{})

	err := metricsMemoryRepo.Ping()
	require.NoError(t, err)

	var metricValueDelta1 int64 = 11
	var metricValueDelta2 int64 = 22
	var metricValueDelta3 float64 = 27.5

	metricValueList := map[string]MetricValue{
		"PollCount1": {
			MType: MeticTypeCounter,
			Delta: &metricValueDelta1,
		},
		"PollCount2": {
			MType: MeticTypeCounter,
			Delta: &metricValueDelta2,
		},
		"Gauge1": {
			MType: MeticTypeGauge,
			Value: &metricValueDelta3,
		},
	}

	err = metricsMemoryRepo.UpdateMany(metricValueList)
	require.NoError(t, err)

	_, err = metricsMemoryRepo.Read("PollCount1", MeticTypeCounter)
	require.NoError(t, err)

	_, err = metricsMemoryRepo.Read("PollCount2", MeticTypeCounter)
	require.NoError(t, err)

	_, err = metricsMemoryRepo.Read("Gauge1", MeticTypeGauge)
	require.NoError(t, err)

	err = metricsMemoryRepo.Close()
	require.NoError(t, err)
}

func TestMemoryRepoUpdateManySlice(t *testing.T) {
	metricsMemoryRepo := NewMetricsMemoryRepo(config.StoreConfig{})

	err := metricsMemoryRepo.Ping()
	require.NoError(t, err)

	var metricValueDelta1 int64 = 11
	var metricValueDelta2 int64 = 22
	var metricValueDelta3 float64 = 27.5

	metricValueList := []Metric{
		{
			ID: "PollCount1",
			MetricValue: MetricValue{
				MType: MeticTypeCounter,
				Delta: &metricValueDelta1,
			},
		},
		{
			ID: "PollCount2",
			MetricValue: MetricValue{
				MType: MeticTypeCounter,
				Delta: &metricValueDelta2,
			},
		},
		{
			ID: "Gauge1",
			MetricValue: MetricValue{
				MType: MeticTypeGauge,
				Value: &metricValueDelta3,
			},
		},
	}

	err = metricsMemoryRepo.UpdateManySliceMetric(metricValueList)
	require.NoError(t, err)

	_, err = metricsMemoryRepo.Read("PollCount1", MeticTypeCounter)
	require.NoError(t, err)

	_, err = metricsMemoryRepo.Read("PollCount2", MeticTypeCounter)
	require.NoError(t, err)

	_, err = metricsMemoryRepo.Read("Gauge1", MeticTypeGauge)
	require.NoError(t, err)

	err = metricsMemoryRepo.Close()
	require.NoError(t, err)
}
