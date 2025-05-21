// Application entry point

package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tg_app_micserv/internal/config"
	"tg_app_micserv/internal/kafka/consumer"
	"tg_app_micserv/internal/kafka/producer"
	"tg_app_micserv/internal/tg_session_storage"
	"tg_app_micserv/tools/logger"
)

const (
	pathEnv = ".env"
	lbl     = "tg_app_micserv/cmd/main.go/main()"
)

func main() {
	// Инициализация логера
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)
	slog.Info("Запустили обертку logger")

	err := godotenv.Load(pathEnv)
	if err != nil {
		slog.Error("Неудалось найти файл .env", "error", err)
		log.Fatal(err)
	}
	slog.Info("Подгрузили файл .env")

	// Загрузка конфигов
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Не удалось загрузить конфиг", "error", err)
		log.Fatal(err)
	}
	slog.Info("Успешно создали объект конфига")

	// Инициализация зависимостей
	// Инициализация хранилища сессии Telegram-клиента (локальный файл session.json)
	sessionStorage := tg_session_storage.NewFileSessionStorage("session.json")
	slog.Info("Успешно создали объект хранилища сессии")

	// Создание нового Telegram клиента с передачей API ID, Hash, телефона и 2FA пароля
	//tgClient := tg_init_parser.NewClient(cfg.API_ID, cfg.API_Hash, cfg.Phone, cfg.Two_F_Password, sessionStorage)
	//slog.Info("Успешно создали объект Telegram клиента")

	// Создание сервиса для получения и очистки сообщений, использующего Telegram клиента
	//serviceParser := service_parser.NewServiceParser(tgClient)
	//slog.Info("Успешно создали Сервис для парсинга постов")

	// Создание обработчика HTTP-запросов, передающего в него сервис парсер постов
	//messageHandler := handlers.NewMessageHandler(serviceParser)
	//slog.Info("Успешно создали объект обработчика HTTP-запросов")

	// Создание kafka Producer-а
	kafkaProducer, err := producer.NewProducer(cfg)
	if err != nil {
		slog.Error("Не удалось создать Kafka-продюсер", "error", err)
		log.Fatal(err)
	}
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			slog.Error("Ошибка при закрытии Kafka-Producer", "error", err)
		}
	}()
	slog.Info("Успешно создали kafka Producer-а")

	// Создание kafka Consumer-а
	kafkaConsumer := consumer.NewConsumer(cfg)
	defer func() {
		if err := kafkaConsumer.Close(); err != nil {
			slog.Error("Ошибка при закрытии Kafka-Consumer", "error", err)
		}
	}()
	slog.Info("Успешно создали kafka Consumer-а")
	ctx, cancel := context.WithCancel(context.Background())
	// Запуск консьюмера
	go func() {
		if err := kafkaConsumer.Consume(ctx, cfg, sessionStorage); err != nil {
			log.Fatalf("Ошибка в работе консьюмера: %v", err)
		}
	}()

	// Создание сервера
	//srv := server.NewServer(cfg.Port, messageHandler)
	//slog.Info("Успешно создали объект сервер")

	// Канал для сигналов прерывания
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	//go func() {
	//	if err := srv.Run(); err != nil {
	//		slog.Error("Сервер не запустился", "error", err)
	//		log.Fatal(err)
	//	}
	//}()

	// Ожидание сигнала прерывания
	<-stop
	slog.Info("Получен сигнал завершения, инициируется мягкая остановка")

	cancel()
	if err := kafkaConsumer.Close(); err != nil {
		log.Fatalf("Ошибка закрытия консьюмера: %v", err)
	}
	slog.Info("Успешно закрыли консьюмера")

	// Создание контекста с таймаутом для graceful shutdown
	//ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	// Завершение работы сервера
	//if err := srv.Shutdown(ctx1); err != nil {
	//	slog.Error("Ошибка при остановке сервера", "error", err)
	//}

	// Закрытие Telegram-клиента
	//slog.Info("Завершение Telegram-клиента")
	//if err := tgClient.Close(ctx1); err != nil {
	//	slog.Error("Ошибка при закрытии Telegram-клиента", "error", err)
	//}

	//slog.Info("Сервер успешно остановлен")
}
