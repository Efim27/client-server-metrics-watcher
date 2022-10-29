//go:build integration
// +build integration

package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"
	"metrics/internal/server/config"
)

const TempDBRepoFilePath = "tempDBRepoFile"

type MetricsDBRepoSuite struct {
	suite.Suite
	metricsRepo        *DBRepo
	db                 *sql.DB
	cleaner            dbcleaner.DbCleaner
	testingContainerDB *dockertest.Resource
	testingPoolDB      *dockertest.Pool
	repoFile           *os.File
}

func (suite *MetricsDBRepoSuite) SetupSuite() {
	var err error

	suite.testingPoolDB, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	log.Println("TestingDB container starting...")
	suite.testingContainerDB, err = suite.testingPoolDB.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_DB=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=Ug6v3NkkE623",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	containerHostPort := suite.testingContainerDB.GetHostPort("5432/tcp")
	DSN := fmt.Sprintf("postgres://postgres:Ug6v3NkkE623@%s/postgres?sslmode=disable", containerHostPort)

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	log.Println("TestingDB container is started")

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	var metricsRepo DBRepo
	err = suite.testingPoolDB.Retry(func() error {
		metricsRepo, err = NewDBRepo(config.StoreConfig{
			File:        TempDBRepoFilePath,
			DatabaseDSN: DSN,
		})
		if err != nil {
			log.Println(err)
			return err
		}

		log.Println(metricsRepo.Ping())
		return metricsRepo.Ping()
	})
	suite.NoError(err)
	suite.metricsRepo = &metricsRepo
	suite.db = metricsRepo.DB()

	err = suite.metricsRepo.InitTables()
	suite.NoError(err)

	suite.repoFile, err = os.OpenFile(TempDBRepoFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	suite.NoError(err)
	suite.metricsRepo.InitFromFile()

	cleanerEngine := engine.NewPostgresEngine(DSN)
	suite.cleaner = dbcleaner.New()
	suite.cleaner.SetEngine(cleanerEngine)
}

func (suite *MetricsDBRepoSuite) TearDownSuite() {
	defer func() {
		err := suite.testingPoolDB.Purge(suite.testingContainerDB)
		if err != nil {
			log.Println(err)
		}
	}()

	err := suite.metricsRepo.Save()
	suite.NoError(err)

	err = suite.metricsRepo.Close()
	suite.NoError(err)

	err = suite.cleaner.Close()
	suite.NoError(err)

	err = suite.repoFile.Close()
	suite.NoError(err)
	err = os.Remove(TempDBRepoFilePath)
	suite.NoError(err)
}

func (suite *MetricsDBRepoSuite) SetupTest() {
	suite.cleaner.Acquire("counter")
	suite.cleaner.Acquire("gauge")
}

func (suite *MetricsDBRepoSuite) TearDownTest() {
	suite.cleaner.Clean("counter")
	suite.cleaner.Clean("gauge")
}

func (suite *MetricsDBRepoSuite) TestDBRepo_Ping() {
	err := suite.metricsRepo.Ping()
	suite.NoError(err)
}

func (suite *MetricsDBRepoSuite) TestDBRepo_ReadEmpty() {
	err := suite.metricsRepo.Ping()
	suite.NoError(err)

	_, err = suite.metricsRepo.Read("PollCount", MeticTypeCounter)
	suite.Error(err)

	_, err = suite.metricsRepo.Read("gauge", MeticTypeGauge)
	suite.Error(err)
}

func (suite *MetricsDBRepoSuite) TestDBRepo_ReadWrite() {
	err := suite.metricsRepo.Ping()
	suite.NoError(err)

	var metricValue1 int64 = 7
	err = suite.metricsRepo.Update("PollCount", MetricValue{
		MType: MeticTypeCounter,
		Delta: &metricValue1,
	})
	suite.NoError(err)

	var metricGauge1 float64 = 27.1
	err = suite.metricsRepo.Update("Gauge", MetricValue{
		MType: MeticTypeGauge,
		Value: &metricGauge1,
	})
	suite.NoError(err)

	metricValueCounter, err := suite.metricsRepo.Read("PollCount", MeticTypeCounter)
	suite.NoError(err)
	suite.EqualValues(metricValue1, *metricValueCounter.Delta)

	metricValueGauge, err := suite.metricsRepo.Read("Gauge", MeticTypeGauge)
	suite.NoError(err)
	suite.EqualValues(metricGauge1, *metricValueGauge.Value)

	metricValueCounter, err = suite.metricsRepo.readCounter("PollCount")
	suite.NoError(err)
	suite.EqualValues(metricValue1, *metricValueCounter.Delta)

	metricValueGauge, err = suite.metricsRepo.Read("Gauge", MeticTypeGauge)
	suite.NoError(err)
	suite.EqualValues(metricGauge1, *metricValueGauge.Value)
}

func (suite *MetricsDBRepoSuite) TestDBRepo_ReadWriteMany() {
	err := suite.metricsRepo.Ping()
	suite.NoError(err)

	var metricValueRaw1 int64 = 27
	metricValue1 := MetricValue{
		MType: MeticTypeCounter,
		Delta: &metricValueRaw1,
	}

	var metricValueRaw2 float64 = 29.2
	metricGauge1 := MetricValue{
		MType: MeticTypeGauge,
		Value: &metricValueRaw2,
	}

	repoMetricMap := MetricMap{"Counter1": metricValue1, "Gauge1": metricGauge1}
	err = suite.metricsRepo.UpdateMany(repoMetricMap)
	suite.NoError(err)

	repoCounterMap, err := suite.metricsRepo.readAllCounter()
	suite.NoError(err)
	suite.EqualValues(MetricMap{"Counter1": metricValue1}, repoCounterMap)

	repoGaugeMap, err := suite.metricsRepo.readAllGauge()
	suite.NoError(err)
	suite.EqualValues(MetricMap{"Gauge1": metricGauge1}, repoGaugeMap)

	repoAllMetricsMap := suite.metricsRepo.ReadAll()
	suite.EqualValues(MetricMap{"Counter1": metricValue1}, repoAllMetricsMap[MeticTypeCounter])
	suite.EqualValues(MetricMap{"Gauge1": metricGauge1}, repoAllMetricsMap[MeticTypeGauge])
}

func TestUploaderSuite(t *testing.T) {
	suite.Run(t, new(MetricsDBRepoSuite))
}
