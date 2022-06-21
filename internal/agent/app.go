package agent

import (
	"log"
	"os"
	"syscall"
	"time"

	"metrics/internal/agent/config"
	"metrics/internal/agent/metricsuploader"
	"metrics/internal/agent/statsreader"
)

type AppHTTP struct {
	isRun   bool
	timeLog struct {
		startTime       time.Time
		lastRefreshTime time.Time
		lastUploadTime  time.Time
	}
	metricsUplader *metricsuploader.MetricsUplader
	config         config.Config
}

func NewHTTPClient(config config.Config) *AppHTTP {
	var app AppHTTP
	app.config = config
	app.metricsUplader = metricsuploader.NewMetricsUploader(app.config.HTTPClientConnection)

	return &app
}

func (app *AppHTTP) Run() {
	var memStatistics statsreader.MemoryStatsDump
	signalChanel := make(chan os.Signal, 1)

	app.timeLog.startTime = time.Now()
	app.isRun = true

	tickerStatisticsRefresh := time.NewTicker(app.config.PollInterval)
	tickerStatisticsUpload := time.NewTicker(app.config.ReportInterval)

	for app.isRun {
		select {
		case timeTickerRefresh := <-tickerStatisticsRefresh.C:
			log.Println("Refresh")
			app.timeLog.lastRefreshTime = timeTickerRefresh
			memStatistics.Refresh()
		case timeTickerUpload := <-tickerStatisticsUpload.C:
			app.timeLog.lastUploadTime = timeTickerUpload
			log.Println("Upload")

			err := app.metricsUplader.MemoryStatsUpload(memStatistics)
			if err != nil {
				log.Println("Error!")
				log.Println(err)

				app.Stop()
			}
		case osSignal := <-signalChanel:
			switch osSignal {
			case syscall.SIGTERM:
				log.Println("syscall: SIGTERM")
			case syscall.SIGINT:
				log.Println("syscall: SIGINT")
			case syscall.SIGQUIT:
				log.Println("syscall: SIGQUIT")
			}
			app.Stop()
		}
	}
}

func (app *AppHTTP) Stop() {
	app.isRun = false
}

func (app *AppHTTP) IsRun() bool {
	return app.isRun
}
