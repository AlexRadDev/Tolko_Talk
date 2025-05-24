package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"text_to_speech_app/internal/text_to_audio/google_text_to_audio"
	"text_to_speech_app/internal/text_to_audio/interface_text_to_audio"

	"text_to_speech_app/internal/kafka/consumer"
	"text_to_speech_app/internal/kafka/producer"
	//"text_to_speech_app/internal/server"
	"text_to_speech_app/internal/config"
	"text_to_speech_app/tools/logger"
)

const (
	envPath = ".env"
)

func main() {
	const lblmain = "cmd/main.go/main()"
	myLogger := logger.NewColorLogger(lblmain)
	slog.SetDefault(myLogger)

	// Загружаем переменные окружения из файла .env
	//if err := godotenv.Load(envPath); err != nil {
	//	slog.Error("Ошибка загрузки .env файла", slog.Any("error", err))
	//	os.Exit(1)
	//}
	//slog.Info("Успешно прочитали файл .env")

	// Загружаем конфигурацию из переменных окружения
	cfg, err := config.Load()
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка загрузки конфигурации", err))
		os.Exit(1)
	}
	slog.Info("Успешно создали объект конфига")

	// Создаём слой бизнес-логики
	GoogleTextToAudio := google_text_to_audio.NewGoogleTextToAudio(cfg)
	slog.Info("Успешно создали объект Text-to-Speech сервиса")

	// Объявили интерфейс обработки текста в аудио
	var аudioSynthesizer interface_text_to_audio.InterfaceTextToAudio
	// Присвоили интерфейсу реализацию через Google Text to Speech
	аudioSynthesizer = GoogleTextToAudio

	// Инициализируем Kafka-продюсер
	kafkaProducer := producer.NewProducer(cfg)
	slog.Info("Успешно создали Kafka-продюсер")

	// Инициализируем Kafka-консьюмер
	kafkaConsumer := consumer.NewConsumer(cfg, kafkaProducer, аudioSynthesizer)
	slog.Info("Успешно создали Kafka-консьюмер")

	// Создаём HTTP-сервер, внедряя бизнес-логику
	//srv := server.NewServer(cfg.ServerPort, GoogleTextToAudio)
	//slog.Info("Успешно создали HTTP-сервер")

	// Создаём канал для получения сигналов ОС (для graceful shutdown)
	quit := make(chan os.Signal, 1)

	// Регистрируем сигналы SIGINT и SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем Kafka-консьюмер в отдельной горутине
	go func() {
		slog.Info("Запуск Kafka-консьюмера")
		if err := kafkaConsumer.Consume(context.Background(), cfg); err != nil {
			slog.Error("Ошибка работы Kafka-консьюмера", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Запускаем HTTP-сервер в отдельной горутине
	//go func() {
	//	slog.Info("Запуск HTTP-сервера", slog.String("address", cfg.ServerPort))
	//	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	//		slog.Error("Ошибка работы сервера", slog.Any("error", err))
	//		os.Exit(1)
	//	}
	//}()

	// Ожидаем сигнал завершения
	<-quit
	slog.Info("Получен сигнал завершения, инициируем graceful shutdown")

	// Создаём контекст с таймаутом для graceful shutdown
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	// Закрываем Kafka-консьюмер
	if err := kafkaConsumer.Close(); err != nil {
		slog.Error("Ошибка закрытия Kafka-консьюмера", slog.Any("error", err))
	}

	// Закрываем Kafka-продюсер
	if err := kafkaProducer.Close(); err != nil {
		slog.Error("Ошибка закрытия Kafka-продюсера", slog.Any("error", err))
	}

	// Выполняем graceful shutdown сервера
	//if err := srv.Shutdown(ctx); err != nil {
	//	slog.Error("Ошибка при завершении работы сервера", slog.Any("error", err))
	//	os.Exit(1)
	//}

	slog.Info("Микросервис text_to_speech_micserv успешно завершил работу")
}
