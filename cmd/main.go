package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"tolko_talk/internal/app_telega"
	"tolko_talk/internal/app_text_to_speech"
)

const (
	pathEnv      = "D:/go_progect_for_Git/tolko_talk/.env"
	timeNews     = 20 // Промежуток времени в минутах (за который нужно скачать новости)
	mp3Paht      = "sound_001.mp3"
	SpeakingRate = 1.5 // От 0.2 до 4.0
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

	//textNews := "Welcome to Tbilisi, the vibrant capital of Georgia! Nestled along the Mtkvari River, this city blends ancient history with modern charm. Explore the cobblestone streets of the Old Town, visit the stunning Holy Trinity Cathedral, or enjoy a glass of local wine at a cozy café. Tbilisi awaits you with warmth and wonder."
	//textNews = "Текст: Восстанавливаемый петербургскими специалистами Драматический театр в Мариуполе начнет работать в этом году, сообщил губернатор Беглов.\nМариупольцы опасались, что легендарный театр после страшного взрыва, устроенного нацистами, потерян навсегда. Но нашим строителям, за что мы их благодарим, удалось не просто восстановить здание, но и наполнить его лучшим современным оборудованием. Беглов недавно сообщал, что готовность здания составляет порядка 70%. А теперь рассказал, что Петербургский институт сценических искусств открыл целевую программу для подготовки актеров и режиссеров из новых регионов, которые придут работать в восстановленный театр."

	// Запускаем перевод текста в аудио
	if err := app_text_to_speech.SynthesizeText(textNews, keyToSpeech, mp3Paht, SpeakingRate); err != nil {
		log.Fatalf("Ошибка функции SynthesizeText: %v", err)
	}
}
