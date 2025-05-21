package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text_to_speech_app/internal/app_text_to_speech"
	"text_to_speech_app/internal/kafka/consumer"
	"text_to_speech_app/internal/kafka/producer"
	"text_to_speech_app/internal/server"
	"time"

	"github.com/joho/godotenv"

	"text_to_speech_app/internal/config"
	"text_to_speech_app/tools/logger"
)

const (
	envPath = ".env"
)

// Функция main — точка входа приложения
func main() {
	const lblmain = "cmd/main.go/main()"
	myLogger := logger.NewColorLogger(lblmain)
	slog.SetDefault(myLogger)

	// Загружаем переменные окружения из файла .env
	if err := godotenv.Load(envPath); err != nil {
		slog.Error("Ошибка загрузки .env файла", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Успешно прочитали файл .env")

	// Загружаем конфигурацию из переменных окружения
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Успешно создали объект конфига")

	// Инициализируем Kafka-продюсер
	kafkaProducer := producer.NewProducer([]string{cfg.KafkaPort}, cfg.NameTopicProdus)
	slog.Info("Успешно создали Kafka-продюсер", slog.String("topic", cfg.NameTopicProdus))

	// Инициализируем Kafka-консьюмер
	kafkaConsumer := consumer.NewConsumer([]string{cfg.KafkaPort}, cfg.NameTopicProdus, cfg.KafkaGroupID)
	slog.Info("Успешно создали Kafka-консьюмер", slog.String("topic", cfg.NameTopicProdus))

	// Создаём слой бизнес-логики
	ttsService := app_text_to_speech.NewService(cfg.GoogleCredentialsFile, kafkaProducer)
	slog.Info("Успешно создали объект Text-to-Speech сервиса")

	// Создаём HTTP-сервер, внедряя бизнес-логику
	srv := server.NewServer(cfg.ServerPort, ttsService)
	slog.Info("Успешно создали HTTP-сервер")

	// Создаём канал для получения сигналов ОС (для graceful shutdown)
	quit := make(chan os.Signal, 1)

	// Регистрируем сигналы SIGINT и SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем Kafka-консьюмер в отдельной горутине
	go func() {
		slog.Info("Запуск Kafka-консьюмера")
		if err := kafkaConsumer.Consume(context.Background(), ttsService); err != nil {
			slog.Error("Ошибка работы Kafka-консьюмера", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Запускаем HTTP-сервер в отдельной горутине
	go func() {
		slog.Info("Запуск HTTP-сервера", slog.String("address", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка работы сервера", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Ожидаем сигнал завершения
	<-quit
	slog.Info("Получен сигнал завершения, инициируем graceful shutdown")

	// Создаём контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Закрываем Kafka-консьюмер
	if err := kafkaConsumer.Close(); err != nil {
		slog.Error("Ошибка закрытия Kafka-консьюмера", slog.Any("error", err))
	}

	// Закрываем Kafka-продюсер
	if err := kafkaProducer.Close(); err != nil {
		slog.Error("Ошибка закрытия Kafka-продюсера", slog.Any("error", err))
	}

	// Выполняем graceful shutdown сервера
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Ошибка при завершении работы сервера", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Сервер успешно завершил работу")
}
