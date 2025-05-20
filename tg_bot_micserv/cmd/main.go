package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
	"tg_bot/internal/server"

	"tg_bot/internal/config"
	"tg_bot/internal/kafka"
	"tg_bot/internal/repo_user_requests"
	"tg_bot/internal/tg_bot_init"
	tg_bot_router2 "tg_bot/internal/tg_bot_router"
	"tg_bot/internal/tg_bot_user_case"
	"tg_bot/tools/logger"
)

const (
	envPath = ".env"
)

func main() {
	// Инициализируем логгер с меткой
	const lblmain = "cmd/main.go/main()"
	log := logger.NewColorLogger(lblmain)
	slog.SetDefault(log)

	// Загружаем переменные окружения из файла .env
	if err := godotenv.Load(envPath); err != nil {
		slog.Error("Ошибка загрузки .env файла", "error", err)
		os.Exit(1)
	}
	slog.Info("Успешно прочитали файл .env")

	// Загружаем конфигурацию из переменных окружения
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}
	slog.Info("Успешно создали объект кофига")

	// Инициализируем Telegram-бот с токеном из конфигурации
	tgBot, err := tg_bot_init.NewBot(cfg.TGBotToken)
	if err != nil {
		slog.Error("Ошибка инициализации Telegram-бота", "error", err)
		os.Exit(1)
	}
	slog.Info(fmt.Sprintf("Бот авторизован как @%s", tgBot.Self.UserName))

	// Создаём in-memory репозиторий для хранения состояния пользователей
	repo := repo_user_requests.NewRepoUserRequests()
	slog.Info("Успешно создали хранилище для состояния пользователей")

	// Инициализируем Kafka-продюсер
	kafkaProducer := kafka.NewProducer(cfg.KafkaPort, cfg.NameTopicKafka) //[]string{cfg.KafkaPort}
	slog.Info(fmt.Sprintf("Успешно создали Kafka-продюсер, Name Topic: %v", cfg.NameTopicKafka))

	// Создаём слой бизнес-логики, внедряя репозиторий
	userCase := tg_bot_user_case.NewUseCase(repo, kafkaProducer)
	slog.Info("Успешно создали объект userCase")

	// Создаём HTTP-роутер, внедряя бот и бизнес-логику
	router := tg_bot_router2.NewRouter(tgBot, userCase)

	// Настраиваем Long Polling для получения обновлений
	updates := tg_bot_init.SetupLongPolling(tgBot)
	slog.Info("Запущен режим Long Polling для получения обновлений")

	// Запускаем обработку обновлений в горутине
	go func() {
		for update := range updates {
			router.ProcessUpdate(update)
		}
	}()

	// Настраиваем HTTP-мультиплексор для обработки запросов
	mux := http.NewServeMux() // эта строка Создаёт новый HTTP-мультиплексор (роутер) для обработки HTTP-запросов и сохраняет его в переменную mux.
	// ServeMux — это структура, которая используется для маршрутизации HTTP-запросов.
	// Она сопоставляет (т.е. соединяет) URL-пути (например, /tgBotPost) с обработчиками (функциями, которые обрабатывают эти запросы).
	// Без мультиплексора сервер мог бы обрабатывать только один обработчик для всех запросов.
	// ServeMux позволяет разделить логику обработки в зависимости от пути (много разных вариантов)

	mux.HandleFunc("/tgBotPost", router.HandleUpdate)

	// Создаём HTTP-сервер
	srv := server.NewServer(cfg, mux)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := srv.Start(); err != nil {
			os.Exit(1)
		}
	}()

	// Ожидаем сигнал завершения и выполняем graceful shutdown
	srv.WaitForShutdown()
}
