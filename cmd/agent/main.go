package main

import (
	"metrics/internal/agent"
	"metrics/internal/agent/config"
)

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
	appConfig := config.LoadConfig()
	app := agent.NewAppHTTP(appConfig)
	app.Run()
}
