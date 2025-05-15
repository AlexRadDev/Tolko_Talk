package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"

	"github.com/joho/godotenv"

	//"tolko_talk/internal/app_telega"
	//"tolko_talk/internal/app_text_to_speech"
	"tolko_talk/internal/config"
	"tolko_talk/internal/server"
	"tolko_talk/internal/tg_bot_handler"
	"tolko_talk/internal/tg_bot_init"
	"tolko_talk/internal/tg_bot_usecase"
	"tolko_talk/tools/logger"
)

const (
	pathEnv = "D:/go_progect_for_Git/tolko_talk/.env"
)

func main() {
	const lbl = "cmd/main.go/main()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	slog.Info("Запустили обертку logger")

	// Загружаем файл .env
	err := godotenv.Load(pathEnv)
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}
	slog.Info("Прочитали файл .env")

	// Инициализируем объект config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка чтения .env файла: %v", err)
	}
	slog.Info("Создали объект config")

	// Инициализируем объект бота
	tgBotClient, err := tg_bot_init.NewBot(cfg.TG_Bot_Token)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}
	slog.Info("Создали объект tgBotClient")

	// Инициализируем объект бизнес-логика
	tgBotUseCase := tg_bot_usecase.NewUseCase()
	slog.Info("Создали объект tgBotUseCase")

	// Инициализируем объект tgBotHandler
	tgBotHandler := tg_bot_handler.NewHandler(tgBotUseCase) // передаём usecase в хендлеры
	slog.Info("Создали объект tgBotHandler")

	// Настраиваем получение обновлений через GetUpdates
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := tgBotClient.GetUpdatesChan(updateConfig)
	slog.Info("Бот запущен и ожидает обновления")

	// Запуск сервера
	go server.StartServer(cfg, tgBotClient)

	// Обрабатываем обновления
	for update := range updates {
		if update.Message != nil {
			slog.Info("Получено сообщение", "text", update.Message.Text, "chatID", update.Message.Chat.ID)
			tgBotHandler.HandleMessage(tgBotClient, update.Message)
		}
	}
}

//region Удаляем старый вебхук, если он был
//_, err = tgBotClient.Request(tgbotapi.NewWebhook(""))
//if err != nil {
//	slog.Warn("Не удалось удалить вебхук", "error", err)
//}
//endregion
