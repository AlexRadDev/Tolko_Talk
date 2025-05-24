// curl https://api.telegram.org/bot<7964767527:AAHWj-x5ItudN-IyFEFwPaErz8yCa9y4Elw>/getMe

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"tg_bot/internal/config"
	"tg_bot/internal/kafka/consumer"
	"tg_bot/internal/kafka/producer"
	"tg_bot/internal/repo_user_requests"
	"tg_bot/internal/server"
	"tg_bot/internal/storage"
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
	myLoggeg := logger.NewColorLogger(lblmain)
	slog.SetDefault(myLoggeg)

	//Загружаем переменные окружения из файла .env
	//if err := godotenv.Load(envPath); err != nil {
	//	slog.Error("Ошибка загрузки .env файла", "error", err)
	//	os.Exit(1)
	//}
	//slog.Info("Успешно прочитали файл .env")

	// Загружаем конфигурацию из переменных окружения
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", "error", err)
		//os.Exit(1)
	}
	slog.Info("Успешно создали объект кофига")

	slog.Info(fmt.Sprintf("Токен равен: %v", cfg.TGBotToken))

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

	ctx, cancel := context.WithCancel(context.Background())

	// Инициализируем Kafka-продюсер
	kafkaProducer := producer.NewProducer(cfg)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			slog.Error("Ошибка при закрытии Kafka-продюсера", "error", err)
		} else {
			slog.Info("Kafka-продюсер успешно закрыт")
		}
	}()
	slog.Info("Успешно создали Kafka-продюсер")

	// Создание Топиков в kafka
	for _, nameTopic := range cfg.KafkaTopicSlise {
		kafkaProducer.CheckingAndCreatingTopics(ctx, nameTopic)
	}

	// Инициализируем Kafka-консьюмер 01
	kafkaConsumer01 := consumer.NewConsumer(cfg, 1)
	defer func() {
		if err := kafkaConsumer01.Close(); err != nil {
			slog.Error("Ошибка при закрытии Kafka-консьюмера 01", "error", err)
		} else {
			slog.Info("Kafka-консьюмер 01 успешно закрыт")
		}
	}()
	slog.Info("Успешно создали kafka Consumer-а 01 (text_to_speech → tg_bot)")
	// Запуск консьюмера 01
	go func() {
		if err := kafkaConsumer01.Consume01(ctx, tgBot); err != nil {
			log.Fatalf("Ошибка в работе консьюмера 01: %v", err)
		}
	}()

	kafkaConsumer02 := consumer.NewConsumer(cfg, 2)
	defer func() {
		if err := kafkaConsumer02.Close(); err != nil {
			slog.Error("Ошибка при закрытии Kafka-консьюмера 02", "error", err)
		} else {
			slog.Info("Kafka-консьюмер 02 успешно закрыт")
		}
	}()
	slog.Info("Успешно создали kafka Consumer-а 02 (tg_app → tg_bot)")
	// Запуск консьюмера 02
	go func() {
		if err := kafkaConsumer02.Consume02(ctx, tgBot); err != nil {
			log.Fatalf("Ошибка в работе консьюмера 02: %v", err)
		}
	}()
	defer cancel()

	MapStorage := storage.NewMapStorage()
	slog.Info("Успешно создали объект MapStorage")
	var storageInterface storage.StorageInterface
	storageInterface = MapStorage

	// Создаём слой бизнес-логики, внедряя репозиторий
	userCase := tg_bot_user_case.NewUseCase(repo, kafkaProducer, storageInterface)
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

	// Отменяем контекст для завершения всех горутин
	cancel()
	slog.Info("Инициировано завершение всех горутин")

	// Даём время на завершение операций (например, завершение обработки сообщений)
	select {
	case <-time.After(5 * time.Second):
		slog.Info("Программа завершена после ожидания завершения горутин")
	}
}
