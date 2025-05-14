package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"tolko_talk/internal/app_text_to_speech"

	"github.com/joho/godotenv"
	"tolko_talk/internal/app_telega"
)

const (
	pathEnv      = "D:/go_progect_for_Git/tolko_talk/.env"
	timeNews     = 60 // Промежуток времени в минутах (за который нужно скачать новости)
	mp3Paht      = "sound_001.mp3"
	SpeakingRate = 1.2 // От 0.2 до 4.0
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
		log.Fatalf("Ошибка при переводе API_ID в int: %v %v", err, apiID)
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
	keyToSpeech := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if keyToSpeech == "" {
		log.Fatal("Ошибка: GOOGLE_APPLICATION_CREDENTIALS не задан в .env")
	}
	//slog.Info(fmt.Sprintf("Мой TWO_FACTOR_AUTH из .env: %s", twoFactorAuth))
	slog.Info("Создали переменные API данных")

	// Запускаем парсинг канала телеги
	textNews, err := app_telega.RunTelegaApp(apiID, apiHash, channelName, phone, password, timeNews)
	if err != nil {
		log.Fatalf("Ошибка функции RunTelegaApp: %v", err)
	}

	// Запускаем перевод текста в аудио
	if err := app_text_to_speech.SynthesizeText(textNews, keyToSpeech, mp3Paht, SpeakingRate); err != nil {
		log.Fatalf("Ошибка функции SynthesizeText: %v", err)
	}
}
