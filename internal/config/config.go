package config

import (
	"errors"
	"log"
	"os"
	"strconv"
)

type Config struct {
	TG_Bot_Token         string
	TG_Bot_WebHost       string
	TG_Api_ID            int
	TG_Api_API_Hash      string
	Chanel_Name          string
	MyHpone_for_App      string
	Two_Factor_Auth      string
	Google_Key_To_Speech string
}

// LoadConfig читает данные из .env и создает объект Config
func LoadConfig() (*Config, error) {

	tgBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if tgBotToken == "" {
		return nil, errors.New("переменная окружения TELEGRAM_BOT_TOKEN пуста")
	}
	tgBotWebHost := os.Getenv("TELEGRAM_BOT_PORT")
	if tgBotWebHost == "" {
		return nil, errors.New("переменная окружения TELEGRAM_BOT_PORT пуста")
	}
	apiIDStr := os.Getenv("API_ID")
	if apiIDStr == "" {
		return nil, errors.New("переменная окружения API_ID пуста")
	}
	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		log.Fatalf("Ошибка при переводе API_ID в int: %v %v", err, apiID)
	}
	apiHash := os.Getenv("API_HASH")
	if apiHash == "" {
		return nil, errors.New("переменная окружения API_HASH пуста")
	}
	channelName := os.Getenv("CHANNEL_NAME")
	if channelName == "" {
		return nil, errors.New("переменная окружения CHANNEL_NAME пуста")
	}
	if len(channelName) > 0 && channelName[0] == '@' { // Убираем @ из имени канала
		channelName = channelName[1:]
	}
	phone := os.Getenv("PHONE")
	if phone == "" {
		return nil, errors.New("переменная окружения PHONE пуста")
	}
	password := os.Getenv("TWO_FACTOR_AUTH")
	if password == "" {
		return nil, errors.New("переменная окружения TWO_FACTOR_AUTH пуста")
	}
	keyToSpeech := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if keyToSpeech == "" {
		return nil, errors.New("переменная окружения GOOGLE_APPLICATION_CREDENTIALS пуста")
	}

	return &Config{
		TG_Bot_Token:         tgBotToken,
		TG_Bot_WebHost:       tgBotWebHost,
		TG_Api_ID:            apiID,
		TG_Api_API_Hash:      apiHash,
		Chanel_Name:          channelName,
		MyHpone_for_App:      phone,
		Two_Factor_Auth:      password,
		Google_Key_To_Speech: keyToSpeech,
	}, nil
}
