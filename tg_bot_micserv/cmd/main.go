package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

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
	kafkaProducer := kafka.NewProducer([]string{cfg.KafkaPort}, cfg.NameTopicKafka)
	slog.Info(fmt.Sprintf("Успешно создали Kafka-продюсер, Name Topic: %v", cfg.NameTopicKafka))

	// Создаём слой бизнес-логики, внедряя репозиторий
	userCase := tg_bot_user_case.NewUseCase(repo, kafkaProducer)
	slog.Info("Успешно создали объект userCase")

	// Создаём HTTP-роутер, внедряя бот и бизнес-логику
	router := tg_bot_router2.NewRouter(tgBot, userCase)

	// Создаём HTTP-сервер
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: http.HandlerFunc(router.HandleUpdate),
	}

	// Создаём канал для получения сигналов ОС (для graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер
	go func() {
		slog.Info("Запуск HTTP-сервера", "address", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка работы сервера", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидаем сигнал завершения
	<-quit
	slog.Info("Получен сигнал завершения, инициируем graceful Shutdown")

	// Создаём контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполняем graceful shutdown сервера
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Ошибка при завершении работы сервера", "error", err)
		os.Exit(1)
	}

	slog.Info("Сервер успешно завершил работу")
}
