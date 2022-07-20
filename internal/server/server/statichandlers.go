package server

import (
	"html/template"
	"net/http"

	"go.uber.org/zap"
)

func (server Server) PrintAllMetricStatic(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(server.config.TemplatesAbsPath + "/index.html")
	if err != nil {
		server.logger.Error("cant parse template", zap.Error(err))
		return
	}

	err = t.Execute(rw, server.storage.ReadAll())
	if err != nil {
		server.logger.Error("cant render template", zap.Error(err))
		return
	}
}
