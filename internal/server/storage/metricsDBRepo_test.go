//go:build integration
// +build integration

package storage

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"
	"metrics/internal/server/config"
)

// testing DB, run docker-compose up before testing
const DSN = "postgres://postgres:Ug6v3NkkE623@localhost:5434/postgres?sslmode=disable"

type MetricsDBRepoSuite struct {
	suite.Suite
	metricsRepo *DBRepo
	db          *sql.DB
	cleaner     dbcleaner.DbCleaner
}

func (suite *MetricsDBRepoSuite) SetupSuite() {
	metricsRepo, err := NewDBRepo(config.StoreConfig{
		DatabaseDSN: DSN,
	})
	suite.NoError(err)
	suite.metricsRepo = &metricsRepo
	suite.db = metricsRepo.DB()

	err = metricsRepo.InitTables()
	suite.NoError(err)

	cleanerEngine := engine.NewPostgresEngine(DSN)
	suite.cleaner = dbcleaner.New()
	suite.cleaner.SetEngine(cleanerEngine)
}

func (suite *MetricsDBRepoSuite) TearDownSuite() {
	err := suite.metricsRepo.Close()
	suite.NoError(err)

	err = suite.cleaner.Close()
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
	metricsRepo, err := NewDBRepo(config.StoreConfig{
		DatabaseDSN: DSN,
	})
	suite.NoError(err)

	err = metricsRepo.Ping()
	suite.NoError(err)

	_, err = metricsRepo.Read("PollCount", MeticTypeCounter)
	suite.Error(err)

	err = metricsRepo.Close()
	suite.NoError(err)
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

	metricValue, err := suite.metricsRepo.Read("PollCount", MeticTypeCounter)
	suite.NoError(err)
	suite.EqualValues(metricValue1, *metricValue.Delta)

	err = suite.metricsRepo.Close()
	suite.NoError(err)
}

func TestUploaderSuite(t *testing.T) {
	suite.Run(t, new(MetricsDBRepoSuite))
}
