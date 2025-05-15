package server

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"tolko_talk/internal/app_telega"
	"tolko_talk/tools/logger"

	"tolko_talk/internal/config"
	"tolko_talk/internal/model/tg_bot_model"
)

func HandleSubmit(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const lbl = "internal/server/server.go/HandleSubmit()"
		logger := logger.NewColorLogger(lbl)
		slog.SetDefault(logger)

		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		var request tg_bot_model.TgBotRequest // var request TgBotRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		slog.Info(fmt.Sprintf("Запрос от TG Bota: %+v\n", request))

		// Убираем @ из имени канала
		if len(request.NameChanel) > 0 && request.NameChanel[0] == '@' {
			request.NameChanel = request.NameChanel[1:]
		}

		// Запускаем парсинг канала телеги
		textNews, err := app_telega.RunTelegaApp(cfg.TG_Api_ID, cfg.TG_Api_API_Hash, request.NameChanel, cfg.MyHpone_for_App, cfg.Two_Factor_Auth, request.TimePeriod)
		if err != nil {
			slog.Error(err.Error())
			log.Fatalf("Ошибка функции RunTelegaApp: %v", err)
		}
		slog.Info(fmt.Sprintf("Текст из канала: %+v\n", textNews))

		// Ответ бэкэнда
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Запрос обработан"))
	}
}

func StartServer(cfg *config.Config) {
	const lbl = "internal/server/server.go/StartServer()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	http.HandleFunc("/tgBotPost", HandleSubmit(cfg))

	slog.Info(fmt.Sprintf("Сервер запущен на порту: %v", cfg.ServerPort))
	if err := http.ListenAndServe(cfg.ServerPort, nil); err != nil {
		slog.Error(fmt.Sprintf("Ошибка запуска сервера: %v", err))
	}
}
