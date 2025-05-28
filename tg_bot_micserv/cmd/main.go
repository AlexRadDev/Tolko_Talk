package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"tg_bot/internal/config"
	"tg_bot/internal/kafka/consumer"
	"tg_bot/internal/kafka/producer"
	"tg_bot/internal/metrics/metrics_prometheus"
	"tg_bot/internal/metrics/metrics_time"
	"tg_bot/internal/server"
	"tg_bot/internal/storage/interface_channel_list_storage"
	"tg_bot/internal/storage/interface_requests_storage"
	"tg_bot/internal/storage/map_requests_storage"
	"tg_bot/internal/storage/postgres_channel_list_storage"
	"tg_bot/internal/tg_bot_init"
	tg_bot_router2 "tg_bot/internal/tg_bot_router"
	"tg_bot/internal/tg_bot_user_case"
	"tg_bot/tools/logger"
)

const (
	envPath = ".env"
)

func main() {
	const lblmain = "[myMicro] cmd/main.go/main()"
	myLoggeg := logger.NewColorLogger(lblmain)
	//slog.SetDefault(myLoggeg)

	//region Загрузка переменных окружения из файла .env
	//Загружаем переменные окружения из файла .env
	//if err := godotenv.Load(envPath); err != nil {
	//	slog.Error("Ошибка загрузки .env файла", "error", err)
	//	os.Exit(1)
	//}
	//slog.Info("Успешно прочитали файл .env")
	//endregion

	// Создание конфига
	cfg, err := config.Load()
	if err != nil {
		myLoggeg.Error(fmt.Sprintf("Ошибка загрузки конфигурации: %v", err))
		os.Exit(1)
	}
	myLoggeg.Info("Успешно создали объект кофига")

	metricsTime := metrics_time.NewMetricsTime()
	metricsPrometheus := metrics_prometheus.NewTGBotMetricsPrometheus()

	// Создание Telegram-бота
	tgBot, err := tg_bot_init.NewBot(cfg.TGBotToken)
	if err != nil {
		myLoggeg.Error(fmt.Sprintf("Ошибка инициализации Telegram-бота: %v", err))
		os.Exit(1)
	}
	myLoggeg.Info(fmt.Sprintf("Бот авторизован как @%s", tgBot.Self.UserName))

	// Создание хранилища для состояния запросов пользователей
	storageReqUsers := map_requests_storage.NewRepoUserRequests()
	myLoggeg.Info("Успешно создали хранилище для состояния пользователей")

	var mapRequests interface_requests_storage.UserRequestRepository
	mapRequests = storageReqUsers

	ctx, cancel := context.WithCancel(context.Background())

	// Создание Kafka-продюсера
	kafkaProducer := producer.NewProducer(cfg)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			myLoggeg.Error(fmt.Sprintf("Ошибка при закрытии Kafka-продюсера: %v", err))
		} else {
			myLoggeg.Info("Kafka-продюсер успешно закрыт")
		}
	}()
	myLoggeg.Info("Успешно создали Kafka-продюсер")

	// Создание Топиков в kafka
	for _, nameTopic := range cfg.KafkaTopicSlise {
		err := kafkaProducer.CheckingAndCreatingTopics(ctx, nameTopic)
		if err != nil {
			myLoggeg.Error(fmt.Sprintf("Ошибка при создании топиков в Kafka: %v", err))
		}
	}

	// Создание Kafka-консьюмера 01
	kafkaConsumer01 := consumer.NewConsumer(cfg, 1, metricsTime, metricsPrometheus)
	defer func() {
		if err := kafkaConsumer01.Close(); err != nil {
			myLoggeg.Error(fmt.Sprintf("Ошибка при закрытии Kafka-консьюмера 01: %v", err))
		} else {
			myLoggeg.Info("Kafka-консьюмер 01 успешно закрыт")
		}
	}()
	myLoggeg.Info("Успешно создали kafka Consumer-а 01 (text_to_speech → tg_bot)")
	// Запуск консьюмера 01
	go func() {
		if err := kafkaConsumer01.Consume01(ctx, tgBot); err != nil {
			myLoggeg.Error(fmt.Sprintf("Ошибка в работе консьюмера 01: %v", err))
			log.Fatalf("Ошибка в работе консьюмера 01: %v", err)
		}
	}()
	kafkaConsumer02 := consumer.NewConsumer(cfg, 2, metricsTime, metricsPrometheus)
	defer func() {
		if err := kafkaConsumer02.Close(); err != nil {
			myLoggeg.Error(fmt.Sprintf("Ошибка при закрытии Kafka-консьюмера 02: %v", err))
		} else {
			myLoggeg.Info("Kafka-консьюмер 02 успешно закрыт")
		}
	}()
	myLoggeg.Info("Успешно создали kafka Consumer-а 02 (tg_app → tg_bot)")
	go func() {
		if err := kafkaConsumer02.Consume02(ctx, tgBot); err != nil {
			myLoggeg.Error(fmt.Sprintf("Ошибка в работе консьюмера 02: %v", err))
			log.Fatalf("Ошибка в работе консьюмера 02: %v", err)
		}
	}()
	defer cancel()

	// Создание подключения к БД
	postgresStorage, err := postgres_channel_list_storage.NewPostgresStorage(cfg)
	if err != nil {
		myLoggeg.Error(fmt.Sprintf("Ошибка подключения к бд Postgres: %v", err))
		log.Fatalf("Ошибка подключения к бд Postgres: %v", err)
	}
	defer postgresStorage.Close()

	// Объявили переменную для Storage interface
	var storageInterface interface_channel_list_storage.StorageInterface
	storageInterface = postgresStorage

	// Создание слой бизнес-логики
	userCase := tg_bot_user_case.NewUseCase(mapRequests, kafkaProducer, storageInterface, metricsTime, metricsPrometheus)
	myLoggeg.Info("Успешно создали объект userCase")

	// Создание HTTP-роутера
	router := tg_bot_router2.NewRouter(tgBot, userCase)

	// Настройка Long Polling для получения обновлений
	updates := tg_bot_init.SetupLongPolling(tgBot)
	myLoggeg.Info("Запущен режим Long Polling для получения обновлений")

	// Запуск обработки обновлений состояния запросов
	go func() {
		for update := range updates {
			router.ProcessUpdate(update)
		}
	}()

	// Настраиваем HTTP-мультиплексор для обработки запросов
	mux := http.NewServeMux()
	mux.HandleFunc("/tgBotPost", router.HandleUpdate)
	mux.Handle("/metrics", promhttp.Handler())

	// Создание HTTP-сервера
	srv := server.NewServer(cfg, mux)

	// Запуска сервера
	go func() {
		if err := srv.Start(); err != nil {
			os.Exit(1)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	srv.WaitForShutdown()

	cancel()
	myLoggeg.Info("Инициировано завершение всех горутин")

	select {
	case <-time.After(5 * time.Second):
		myLoggeg.Info("Программа завершена")
	}
}
