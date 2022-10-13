package agent

import (
	"context"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"metrics/internal/agent/config"
	"metrics/internal/agent/metricsuploader"
	"metrics/internal/agent/statsreader"
	"metrics/internal/logger"
)

type AppHTTP struct {
	isRun          bool
	logFile        *os.File
	logger         *zap.Logger
	metricsUplader *metricsuploader.MetricsUplader
	config         config.Config
}

func mustInitLogger(app *AppHTTP) {
	logLevel := zap.InfoLevel
	if app.config.DebugMode {
		logLevel = zap.DebugLevel
	}

	if app.config.LogFile != "" {
		logFile, err := os.OpenFile(app.config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("cant open log file")
		}

		app.logFile = logFile
	}

	app.logger = logger.InitializeLogger(app.logFile, logLevel)
}

func NewAppHTTP(config config.Config) *AppHTTP {
	app := &AppHTTP{}
	app.config = config
	app.metricsUplader = metricsuploader.NewMetricsUploader(app.config.HTTPClientConnection, app.config.SignKey, app.config.PublicKeyRSA)

	mustInitLogger(app)

	return app
}

func refreshAllMetrics(app *AppHTTP, metricsDump *statsreader.MetricsDump, wgRefresh *sync.WaitGroup) {
	wgRefresh.Add(2)

	go func() {
		defer wgRefresh.Done()
		metricsDump.Refresh()
	}()
	go func() {
		err := metricsDump.RefreshExtra()
		if err != nil {
			app.logger.Error("cant refresh extra metrics", zap.Error(err))
		}

		defer wgRefresh.Done()
	}()
}

func uploadMetrics(app *AppHTTP, metricsDump *statsreader.MetricsDump, wgRefresh *sync.WaitGroup) {
	wgRefresh.Wait()

	go func() {
		err := app.metricsUplader.MetricsUploadBatch(*metricsDump)
		if err != nil {
			app.logger.Error("cant upload metrics", zap.Error(err))
		}
	}()
}

func (app *AppHTTP) Run(ctx context.Context) {
	metricsDump, err := statsreader.NewMetricsDump()
	if err != nil {
		app.logger.Error("cant init metrics handler", zap.Error(err))
		return
	}

	app.logger.Info("start")
	app.isRun = true

	tickerStatisticsRefresh := time.NewTicker(app.config.PollInterval)
	tickerStatisticsUpload := time.NewTicker(app.config.ReportInterval)
	wgRefresh := sync.WaitGroup{}

	for app.isRun {
		select {
		case <-tickerStatisticsRefresh.C:
			app.logger.Info("refresh metrics")
			refreshAllMetrics(app, metricsDump, &wgRefresh)
		case <-tickerStatisticsUpload.C:
			app.logger.Info("upload metrics")
			uploadMetrics(app, metricsDump, &wgRefresh)
		case <-ctx.Done():
			app.logger.Info("upload metrics")
			uploadMetrics(app, metricsDump, &wgRefresh)
			wgRefresh.Wait()

			app.Stop()
		}
	}
}

func (app *AppHTTP) Stop() {
	app.logger.Info("stop")
	app.isRun = false

	app.logFile.Close()
}

func (app *AppHTTP) IsRun() bool {
	return app.isRun
}
