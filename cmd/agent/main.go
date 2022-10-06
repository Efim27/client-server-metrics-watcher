// Клиент для сбора runtime метрик и отправки на сервер
package main

import (
	"fmt"

	"metrics/internal/agent"
	"metrics/internal/agent/config"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

// @Title Client-server metrics
// @Description Сервис сбора и хранения метрик
// @Version 1.0

// @Contact.email efim-02@mail.ru

// @BasePath /
// @Host ultimatestore.io:8080

// @contact.name Efim
// @contact.url https://t.me/hima27
// @contact.email efim-02@mail.ru

// @Tag.name Update
// @Tag.description "Группа запросов обновления метрик"

// @Tag.name Value
// @Tag.description "Группа запросов получения значений метрик"

// @Tag.name Static
// @Tag.description "Группа эндпоинтов со статикой"

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	appConfig := config.LoadConfig()
	app := agent.NewAppHTTP(appConfig)
	app.Run()
}
