package server

import (
	"html/template"
	"log"
	"net/http"
)

// PrintAllMetricStatic
// @Tags Static
// @Summary Metric list
// @ID printAllMetricStatic
// @Produce html
// @Success 200
// @Router / [get]
func (server Server) PrintAllMetricStatic(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(server.config.TemplatesAbsPath + "/index.html")
	if err != nil {
		log.Println("Cant parse template ", err)
		return
	}

	err = t.Execute(rw, server.storage.ReadAll())
	if err != nil {
		log.Println("Cant render template ", err)
		return
	}
}
