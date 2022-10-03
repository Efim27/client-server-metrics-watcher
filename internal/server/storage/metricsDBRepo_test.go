package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"metrics/internal/server/config"
)

// testing DB, run docker-compose up before testing
const DSN = "postgres://postgres:Ug6v3NkkE623@localhost:5434/postgres?sslmode=disable"

func TestDBRepo_InitTables(t *testing.T) {
	metricsRepo, err := NewDBRepo(config.StoreConfig{
		DatabaseDSN: DSN,
	})
	require.NoError(t, err)

	err = metricsRepo.Ping()
	require.NoError(t, err)

	err = metricsRepo.InitTables()
	require.NoError(t, err)

	err = metricsRepo.Close()
	require.NoError(t, err)
}

func TestDBRepo_ReadEmpty(t *testing.T) {
	metricsRepo, err := NewDBRepo(config.StoreConfig{
		DatabaseDSN: DSN,
	})
	require.NoError(t, err)

	err = metricsRepo.Ping()
	require.NoError(t, err)

	_, err = metricsRepo.Read("PollCount", MeticTypeCounter)
	require.Error(t, err)

	err = metricsRepo.Close()
	require.NoError(t, err)
}

func TestDBRepo_ReadWrite(t *testing.T) {
	metricsRepo, err := NewDBRepo(config.StoreConfig{
		DatabaseDSN: DSN,
	})
	require.NoError(t, err)

	err = metricsRepo.Ping()
	require.NoError(t, err)

	var metricValue1 int64 = 7
	err = metricsRepo.Update("PollCount", MetricValue{
		MType: MeticTypeCounter,
		Delta: &metricValue1,
	})
	require.NoError(t, err)

	metricValue, err := metricsRepo.Read("PollCount", MeticTypeCounter)
	require.NoError(t, err)
	require.EqualValues(t, metricValue1, *metricValue.Delta)

	err = metricsRepo.Close()
	require.NoError(t, err)
}
