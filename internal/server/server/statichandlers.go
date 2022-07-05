package server

import (
	"html/template"
	"log"
	"net/http"
)

func (server Server) PrintAllMetricStatic(rw http.ResponseWriter, request *http.Request) {
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
