//go:build integration
// +build integration

package storage

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
	"gopkg.in/khaiql/dbcleaner.v2"
	"gopkg.in/khaiql/dbcleaner.v2/engine"
	"metrics/internal/server/config"
)

type MetricsDBRepoSuite struct {
	suite.Suite
	metricsRepo        *DBRepo
	db                 *sql.DB
	cleaner            dbcleaner.DbCleaner
	testingContainerDB *dockertest.Resource
	testingPoolDB      *dockertest.Pool
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

	err = metricsRepo.InitTables()
	suite.NoError(err)

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
	err := suite.metricsRepo.Ping()
	suite.NoError(err)

	_, err = suite.metricsRepo.Read("PollCount", MeticTypeCounter)
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

	metricValue, err := suite.metricsRepo.Read("PollCount", MeticTypeCounter)
	suite.NoError(err)
	suite.EqualValues(metricValue1, *metricValue.Delta)
}

func TestUploaderSuite(t *testing.T) {
	suite.Run(t, new(MetricsDBRepoSuite))
}
