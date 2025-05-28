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

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"text_to_speech_app/internal/config"
	"text_to_speech_app/internal/kafka/consumer"
	"text_to_speech_app/internal/kafka/producer"
	"text_to_speech_app/internal/metrics/metrics_prometheus"
	"text_to_speech_app/internal/text_to_audio/google_text_to_audio"
	"text_to_speech_app/internal/text_to_audio/interface_text_to_audio"
	"text_to_speech_app/tools/logger"
)

const (
	envPath = ".env"
)

func main() {
	const lblmain = "[myMicro] cmd/main.go/main()"
	myLogger := logger.NewColorLogger(lblmain)

	//region ЗагруЗка файл .env
	//if err := godotenv.Load(envPath); err != nil {
	//	slog.Error("Ошибка загрузки .env файла", slog.Any("error", err))
	//	os.Exit(1)
	//}
	//slog.Info("Успешно прочитали файл .env")
	//endregion

	// Создаем конфиг
	cfg, err := config.Load()
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка загрузки конфигурации", err))
		os.Exit(1)
	}
	myLogger.Info("Успешно создали объект конфига")

	metricsPrometheus := metrics_prometheus.NewTGBotMetricsPrometheus()

	// Создаём слой бизнес-логики
	GoogleTextToAudio := google_text_to_audio.NewGoogleTextToAudio(cfg, metricsPrometheus)
	myLogger.Info("Успешно создали объект Text-to-Speech")

	// Объявили интерфейс обработки текста в аудио
	var audioSynthesizer interface_text_to_audio.InterfaceTextToAudio

	// Присвоили интерфейсу реализацию через Google Text to Speech
	audioSynthesizer = GoogleTextToAudio

	// Создаем Kafka-продюсер
	kafkaProducer := producer.NewProducer(cfg)
	myLogger.Info("Успешно создали Kafka-продюсер")

	// Создаем Kafka-консьюмер
	kafkaConsumer := consumer.NewConsumer(cfg, kafkaProducer, audioSynthesizer, metricsPrometheus)
	myLogger.Info("Успешно создали Kafka-консьюмер")

	// Запускаем Kafka-консьюмер в отдельной горутине
	go func() {
		slog.Info("Запуск Kafka-консьюмера")
		if err := kafkaConsumer.Consume(context.Background(), cfg); err != nil {
			myLogger.Error(fmt.Sprintf("Ошибка работы Kafka-консьюмера: %v", err))
			os.Exit(1)
		}
	}()

	// Канал для сигналов прерывания
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	http.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: nil,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			myLogger.Error(fmt.Sprintf("Ошибка при запуске сервера: %v", err))
		}
	}()
	myLogger.Info(fmt.Sprintf("Сервер запущен на порту: %v", srv.Addr))

	<-stop
	myLogger.Info("Получен сигнал завершения, инициируем graceful shutdown")

	// Закрываем Kafka-консьюмер
	if err := kafkaConsumer.Close(); err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка закрытия Kafka-консьюмера: %v", err))
	}

	// Закрываем Kafka-продюсер
	if err := kafkaProducer.Close(); err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка закрытия Kafka-продюсера: %v", err))
	}

	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершение работы сервера
	if err := srv.Shutdown(ctx1); err != nil {
		myLogger.Error(fmt.Sprintf("Ошибка при остановке сервера: %v", err))
	}

	myLogger.Info("Микросервис text_to_speech успешно завершил работу")
}
