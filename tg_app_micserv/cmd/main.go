// Application entry point

package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"tg_app_micserv/config"
	"tg_app_micserv/internal/handlers"
	"tg_app_micserv/internal/server"
	"tg_app_micserv/internal/service_parser"
	"tg_app_micserv/internal/tg_parser"
	"tg_app_micserv/internal/tg_session_storage"
	"tg_app_micserv/tools/logger"
)

const pathEnv = ".env"

func main() {
	// Инициализация логера
	const lbl = "tg_app_micserv/cmd/main.go/main()"
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
	tgClient := tg_parser.NewClient(cfg.API_ID, cfg.API_Hash, cfg.Phone, cfg.Two_F_Password, sessionStorage)
	slog.Info("Успешно создали объект Telegram клиента")

	// Создание сервиса для получения и очистки сообщений, использующего Telegram клиента
	serviceParser := service_parser.NewServiceParser(tgClient)
	slog.Info("Успешно создали Сервис для парсинга постов")

	// Создание обработчика HTTP-запросов, передающего в него сервис парсер постов
	messageHandler := handlers.NewMessageHandler(serviceParser)
	slog.Info("Успешно создали объект обработчика HTTP-запросов")

	// Создание сервера
	srv := server.NewServer(cfg.Port, messageHandler)
	slog.Info("Успешно создали объект сервер")

	// Канал для сигналов прерывания
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		if err := srv.Run(); err != nil {
			slog.Error("Сервер не запустился", "error", err)
			log.Fatal(err)
		}
	}()

	// Ожидание сигнала прерывания
	<-stop
	slog.Info("Получен сигнал завершения, инициируется мягкая остановка")

	// Создание контекста с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершение работы сервера
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Ошибка при остановке сервера", "error", err)
	}

	// Закрытие Telegram-клиента
	slog.Info("Завершение Telegram-клиента")
	if err := tgClient.Close(ctx); err != nil {
		slog.Error("Ошибка при закрытии Telegram-клиента", "error", err)
	}

	slog.Info("Сервер успешно остановлен")
}
