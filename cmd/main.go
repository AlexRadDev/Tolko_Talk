package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"tolko_talk/internal/app_telega"
)

const (
	pathEnv  = "D:/go_progect_for_Git/tolko_talk/.env"
	timeNews = 20 // промежуток времени в минутах (за который нужно скачать новости)
)

func main() {
	// Загружаем файл .env
	err := godotenv.Load(pathEnv)
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}
	slog.Info("Прочитали файл .env")

	// Получаем переменные окружения
	apiIDStr := os.Getenv("API_ID")
	if apiIDStr == "" {
		log.Fatal("Ошибка: API_ID не задан в .env")
	}
	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		log.Fatalf("Ошибка при переводе API_ID в int: %v", err)
	}
	apiHash := os.Getenv("API_HASH")
	if apiHash == "" {
		log.Fatal("Ошибка: API_HASH не задан в .env")
	}
	channelName := os.Getenv("CHANNEL_NAME")
	if channelName == "" {
		log.Fatal("Ошибка: CHANNEL_USERNAME не задан в .env")
	}
	if len(channelName) > 0 && channelName[0] == '@' { // Убираем @ из имени канала
		channelName = channelName[1:]
	}
	phone := os.Getenv("PHONE")
	if phone == "" {
		log.Fatal("Ошибка: PHONE не задан в .env")
	}
	password := os.Getenv("TWO_FACTOR_AUTH")
	if password == "" {
		log.Fatal("Ошибка: TWO_FACTOR_AUTH не задан в .env")
	}
	//slog.Info(fmt.Sprintf("Мой TWO_FACTOR_AUTH из .env: %s", twoFactorAuth))
	slog.Info("Создали переменные API данных")

	// Запускаем парсинг канала телеги
	app_telega.RunTelegaApp(apiID, apiHash, channelName, phone, password, timeNews)
}
