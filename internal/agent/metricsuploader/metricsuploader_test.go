package metricsuploader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"metrics/internal/agent/config"
	"metrics/internal/agent/statsreader"
	serverCfg "metrics/internal/server/config"
	"metrics/internal/server/server"
	"metrics/internal/server/storage"
)

type UploaderTestingSuite struct {
	suite.Suite
	serverCtx           context.Context
	serverCtxCancel     context.CancelFunc
	metricsUploader     *MetricsUplader
	metricsUploaderGRPC *MetricsUploaderGRPC
}

func (suite *UploaderTestingSuite) SetupSuite() {
	suite.serverCtx, suite.serverCtxCancel = context.WithCancel(context.Background())
	const ServerGRPCAddr = "127.0.0.1:50051"
	serverAPI := server.NewServer(serverCfg.Config{
		ServerAddr:     "127.0.0.1:8080",
		ServerGRPCAddr: ServerGRPCAddr,
	})

	go serverAPI.Run(suite.serverCtx)

	agentConfig := config.LoadConfig()
	suite.metricsUploader = NewMetricsUploader(agentConfig.HTTPClientConnection, "", "")

	clientIP, err := suite.metricsUploader.IP()
	suite.NoError(err)
	suite.NotEmpty(clientIP)

	suite.metricsUploaderGRPC, err = NewMetricsUploaderGRPC(ServerGRPCAddr)
	suite.NoError(err)
}

func (suite *UploaderTestingSuite) TearDownSuite() {
	suite.serverCtxCancel()
}

func (suite *UploaderTestingSuite) TestUploadGRPC() {
	metricsDump, err := statsreader.NewMetricsDump()
	suite.NoError(err)
	metricsDump.Refresh()

	suite.NotNil(metricsDump)
	err = suite.metricsUploaderGRPC.Upload(*metricsDump)
	suite.NoError(err)
}

func (suite *UploaderTestingSuite) TestUploadOne() {
	err := suite.metricsUploader.oneStatUpload(storage.MeticTypeCounter, "Counter1", "27")
	suite.NoError(err)

	err = suite.metricsUploader.oneStatUpload(storage.MeticTypeGauge, "Gauge1", "29.1")
	suite.NoError(err)
}

func (suite *UploaderTestingSuite) TestUploadJSON() {
	metricsDump, err := statsreader.NewMetricsDump()
	suite.NoError(err)
	metricsDump.Refresh()

	suite.NotNil(metricsDump)
	err = suite.metricsUploader.MetricsUploadBatch(*metricsDump)
	suite.NoError(err)
}

func (suite *UploaderTestingSuite) TestUploadAsync() {
	metricsDump, err := statsreader.NewMetricsDump()
	suite.NoError(err)
	metricsDump.Refresh()

	suite.NotNil(metricsDump)
	err = suite.metricsUploader.MetricsUploadAsync(*metricsDump)
	suite.NoError(err)
}

func (suite *UploaderTestingSuite) TestUploadSync() {
	metricsDump, err := statsreader.NewMetricsDump()
	suite.NoError(err)
	metricsDump.Refresh()

	suite.NotNil(metricsDump)
	err = suite.metricsUploader.MetricsUploadSync(*metricsDump)
	suite.NoError(err)
}

func TestUploaderSuite(t *testing.T) {
	suite.Run(t, new(UploaderTestingSuite))
}

func BenchmarkUploader(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	serverAPI := server.NewServer(serverCfg.Config{
		ServerAddr: "127.0.0.1:8080",
	})

	go serverAPI.Run(ctx)
	defer cancel()

	metricsDump, err := statsreader.NewMetricsDump()
	if err != nil {
		b.Error(err)
	}
	metricsDump.Refresh()

	metricsUploader := NewMetricsUploader(config.HTTPClientConfig{
		ServerAddr: "127.0.0.1:8080",
	}, "", "")

	b.Run("sync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = metricsUploader.MetricsUploadSync(*metricsDump)
			if err != nil {
				b.Error(err)
			}
		}
	})

	b.Run("async", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = metricsUploader.MetricsUploadAsync(*metricsDump)
			if err != nil {
				b.Error(err)
			}
		}
	})

	b.Run("JSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = metricsUploader.MetricsUploadBatch(*metricsDump)
			if err != nil {
				b.Error(err)
			}
		}
	})
}
