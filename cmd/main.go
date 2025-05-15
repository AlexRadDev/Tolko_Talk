package main

import (
	//"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"
	//"net/http"
	//"os"

	"github.com/joho/godotenv"

	//"tolko_talk/internal/app_telega"
	//"tolko_talk/internal/app_text_to_speech"
	"tolko_talk/internal/config"
	"tolko_talk/internal/tg_bot_handler"
	"tolko_talk/internal/tg_bot_init"
	//"tolko_talk/internal/tg_bot_router"
	"tolko_talk/internal/tg_bot_usecase"
	"tolko_talk/tools/logger"
)

const (
	pathEnv      = "D:/go_progect_for_Git/tolko_talk/.env"
	timeNews     = 60 // Промежуток времени в минутах (за который нужно скачать новости)
	mp3Paht      = "sound_001.mp3"
	SpeakingRate = 1.2 // От 0.2 до 4.0
)

func main() {
	// Инициализируем глобальный логгер
	slog.SetDefault(logger.NewColorLogger())
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

	// Инициализируем объект tgBotRouter
	//tgBotRouter := tg_bot_router.NewRouter(tgBotClient, tgBotHandler)
	//slog.Info("Создали объект tgBotRouter")
	// Удаляем старый вебхук, если он был
	//_, err = tgBotClient.Request(tgbotapi.NewWebhook(""))
	//if err != nil {
	//	slog.Warn("Не удалось удалить вебхук", "error", err)
	//}

	// Настраиваем получение обновлений через GetUpdates
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := tgBotClient.GetUpdatesChan(updateConfig)
	slog.Info("Бот запущен и ожидает обновления")

	// Обрабатываем обновления
	for update := range updates {
		if update.Message != nil {
			slog.Info("Получено сообщение", "text", update.Message.Text, "chatID", update.Message.Chat.ID)
			tgBotHandler.HandleMessage(tgBotClient, update.Message)
		}
	}

	//http.HandleFunc("/"+cfg.TG_Bot_Token, tgBotRouter.HandleUpdate)
	//slog.Info(fmt.Sprintf("Запуск сервера на: %v\n", cfg.TG_Bot_WebHost))
	//if err := http.ListenAndServe(cfg.TG_Bot_WebHost, nil); err != nil {
	//	slog.Error("Сервер не запустился", "error", err)
	//	os.Exit(1)
	//}
}

//// Запускаем парсинг канала телеги
//textNews, err := app_telega.RunTelegaApp(apiID, apiHash, channelName, phone, password, timeNews)
//if err != nil {
//	log.Fatalf("Ошибка функции RunTelegaApp: %v", err)
//}
//
//// Запускаем перевод текста в аудио
//if err := app_text_to_speech.SynthesizeText(textNews, keyToSpeech, mp3Paht, SpeakingRate); err != nil {
//	log.Fatalf("Ошибка функции SynthesizeText: %v", err)
//}
