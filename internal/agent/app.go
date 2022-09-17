package agent

import (
	"log"
	"os"
	"sync"
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
	app.metricsUplader = metricsuploader.NewMetricsUploader(app.config.HTTPClientConnection, app.config.SignKey)

	return &app
}

func (app *AppHTTP) Run() {
	signalChanel := make(chan os.Signal, 1)
	metricsDump, err := statsreader.NewMetricsDump()
	if err != nil {
		log.Println(err)
		return
	}

	app.timeLog.startTime = time.Now()
	app.isRun = true

	tickerStatisticsRefresh := time.NewTicker(app.config.PollInterval)
	tickerStatisticsUpload := time.NewTicker(app.config.ReportInterval)
	wgRefresh := sync.WaitGroup{}

	for app.isRun {
		select {
		case timeTickerRefresh := <-tickerStatisticsRefresh.C:
			app.timeLog.lastRefreshTime = timeTickerRefresh
			wgRefresh.Add(2)

			go func() {
				defer wgRefresh.Done()
				metricsDump.Refresh()
			}()
			go func() {
				err = metricsDump.RefreshExtra()
				if err != nil {
					log.Println(err)
				}

				defer wgRefresh.Done()
			}()
		case timeTickerUpload := <-tickerStatisticsUpload.C:
			app.timeLog.lastUploadTime = timeTickerUpload
			wgRefresh.Wait()

			go func() {
				err = app.metricsUplader.MetricsUploadBatch(*metricsDump)
				if err != nil {
					log.Println(err)

					app.Stop()
				}
			}()
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
