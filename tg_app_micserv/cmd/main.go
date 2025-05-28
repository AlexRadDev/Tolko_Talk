// Application entry point

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"tg_app_micserv/internal/config"
	"tg_app_micserv/internal/kafka/consumer"
	"tg_app_micserv/internal/kafka/producer"
	"tg_app_micserv/internal/metrics/metrics_prometheus"
	"tg_app_micserv/internal/tg_app_init"
	"tg_app_micserv/internal/tg_app_session_storage"
	"tg_app_micserv/tools/logger"
)

const (
	pathEnv         = ".env"
	lbl             = "[myMicro] tg_app_micserv/cmd/main()"
	pathSessionJson = "/app/session.json" // "/app/session.json" - адрес внутри контейнера. (session.json - адрес тут)
)

func main() {
	mylogger := logger.NewColorLogger(lbl)
	//slog.SetDefault(mylogger)

	//region Загрузка файла конфига .env
	//err := godotenv.Load(pathEnv)
	//if err != nil {
	//	slog.Error("Неудалось найти файл .env", "error", err)
	//	log.Fatal(err)
	//}
	//slog.Info("Подгрузили файл .env")
	//endregion

	// Создание объекта сбора метрик Prometheus
	metricsPrometheus := metrics_prometheus.NewTGBotMetricsPrometheus()

	// Создание объекта конфига
	cfg, err := config.Load()
	if err != nil {
		mylogger.Error(fmt.Sprintf("Не удалось загрузить конфиг: %v", err))
		log.Fatal(err)
	}
	mylogger.Info("Успешно создали объект конфига")

	// Создание хранилища сессии Telegram-клиента (локальный файл session.json)
	sessionStorage := tg_app_session_storage.NewFileSessionStorage(pathSessionJson)
	mylogger.Info("Успешно создали объект хранилища сессии")

	appTGPostParser := tg_app_init.NewTGAppClient(cfg, sessionStorage, pathSessionJson)
	mylogger.Info("Успешно создали tg app клиента")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := appTGPostParser.Start(ctx, cfg.Phone, cfg.Two_F_Password); err != nil {
			mylogger.Error(fmt.Sprintf("Ошибка при запуске Telegram-клиента^ %v", err))
			log.Fatal(err)
		}
	}()
	mylogger.Info("Успешно запустили tg app клиента")

	// Создание kafka Producer-а
	kafkaProducer := producer.NewProducer(cfg)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			mylogger.Error(fmt.Sprintf("Ошибка при закрытии Kafka-Producer: %v", err))
		}
	}()
	mylogger.Info("Успешно создали kafka Producer-а")

	// Создание kafka Consumer-а
	kafkaConsumer := consumer.NewConsumer(cfg, metricsPrometheus, appTGPostParser)
	defer func() {
		if err := kafkaConsumer.Close(); err != nil {
			mylogger.Error(fmt.Sprintf("Ошибка при закрытии Kafka-Consumer: %v", err))
		}
	}()
	mylogger.Info("Успешно создали kafka Consumer-а")

	go func() {
		if err := kafkaConsumer.Consume(ctx, cfg, kafkaProducer); err != nil {
			log.Fatalf("Ошибка в работе консьюмера: %v", err)
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
			mylogger.Error(fmt.Sprintf("Ошибка при запуске сервера: %v", err))
		}
	}()
	mylogger.Info(fmt.Sprintf("Сервер запущен на порту: %v", srv.Addr))

	<-stop
	mylogger.Info("Получен сигнал завершения, инициируется мягкая остановка")

	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := appTGPostParser.Close(ctx1); err != nil {
		mylogger.Error(fmt.Sprintf("Ошибка при закрытии Telegram-клиента: %v", err))
	}
	// Завершение работы сервера
	if err := srv.Shutdown(ctx1); err != nil {
		mylogger.Error(fmt.Sprintf("Ошибка при остановке сервера: %v", err))
	}

	mylogger.Info("Сервер успешно остановлен")
}
