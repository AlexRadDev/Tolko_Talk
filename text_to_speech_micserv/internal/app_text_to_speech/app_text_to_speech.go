// Файл app_text_to_speech.go реализует бизнес-логику микросервиса Text-to-Speech.
// Содержит слой UseCase, который взаимодействует с Google Text-to-Speech API для синтеза речи.

package app_text_to_speech

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"sort"

	"github.com/hajimehoshi/go-mp3"
	"google.golang.org/api/option"
	texttospeech "google.golang.org/api/texttospeech/v1"
	"text_to_speech_app/internal/config"
	"text_to_speech_app/internal/model/model_kafka"
	"text_to_speech_app/tools/logger"
)

// Service представляет сервис Text-to-Speech
type Service struct {
	credentialsFile string // Путь к файлу учетных данных Google Cloud
}

// NewService создаёт новый экземпляр сервиса Text-to-Speech
func NewService(cfg *config.Config) *Service {
	return &Service{
		credentialsFile: cfg.GoogleCredentialsFile,
	}
}

// Synthesize выполняет синтез речи на основе запроса
func (s *Service) Synthesize(ctx context.Context, messageFromTGAPP *model_kafka.KafkaMessageFromTGAPP) (*model_kafka.KafkaRequestForTGAPP, error) {
	const lblSynthesize = "text_to_speech_micserv/internal/app_text_to_speech/app_text_to_speech.go → Synthesize()"
	myLogger := logger.NewColorLogger(lblSynthesize)

	ctx = context.Background()

	// Создаём новый клиент Text-to-Speech
	client, err := texttospeech.NewService(ctx, option.WithCredentialsFile(s.credentialsFile))
	if err != nil {
		myLogger.Error("Не удалось создать клиент Text-to-Speech", slog.Any("error", err))
		return nil, err
	}
	myLogger.Info("Создали клиента Text-to-Speech")

	// Создаём мапу для хранения аудиоданных по идентификатору
	audioDataMap := make(map[int64][]byte)
	// Создаём срез для хранения ключей мапы
	keys := make([]int64, 0, len(messageFromTGAPP.TextForAudio))
	// Заполняем срез ключами
	for k := range messageFromTGAPP.TextForAudio {
		keys = append(keys, k)
	}
	// Сортируем ключи по возрастанию
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	// Логируем количество текстов для обработки
	myLogger.Info("Отсортированы ключи мапы", slog.Int("key_count", len(keys)))

	// Итерируем по отсортированным ключам
	for _, id := range keys {
		text := messageFromTGAPP.TextForAudio[id] // Получаем текст для текущего идентификатора
		myLogger.Info("Синтез речи для текста", slog.Int64("id", id), slog.String("text", text))
		input := &texttospeech.SynthesisInput{ // Определяем входной текст
			Text: text,
		}
		// Настраиваем параметры голоса
		voice := &texttospeech.VoiceSelectionParams{
			LanguageCode: "ru-RU",            // Код языка (русский)
			Name:         "ru-RU-Standard-B", // Имя голоса
			SsmlGender:   "FEMALE",           // Пол голоса
		}
		// Настраиваем параметры аудио
		audioConfig := &texttospeech.AudioConfig{
			AudioEncoding: "MP3",                         // Формат аудио
			SpeakingRate:  messageFromTGAPP.SpeakingRate, // Скорость речи
		}
		// Создаём запрос для синтеза речи
		ttsReq := &texttospeech.SynthesizeSpeechRequest{
			Input:       input,
			Voice:       voice,
			AudioConfig: audioConfig,
		}

		// Выполняем запрос синтеза речи
		resp, err := client.Text.Synthesize(ttsReq).Do()
		if err != nil {
			myLogger.Error("Не удалось синтезировать речь", slog.Int64("id", id), slog.Any("error", err))
			return nil, err
		}
		myLogger.Info("Успешно синтэзировали речь", slog.Int64("id", id))

		// Декодируем base64-строку в байты
		audioData, err := base64.StdEncoding.DecodeString(resp.AudioContent)
		if err != nil {
			myLogger.Error("Не удалось декодировать аудио base64", slog.Int64("id", id), slog.Any("error", err))
			return nil, err
		}
		myLogger.Info("Успешно декодировали аудио base64 в []byte", slog.Int64("id", id))

		// Сохраняем аудиоданные в мапу
		audioDataMap[id] = audioData
	}

	// Создаём буфер для объединённого аудиофайла, 	Итерируем по отсортированным ключам для объединения аудио
	var combinedAudio bytes.Buffer
	for _, id := range keys {
		audioData := audioDataMap[id]

		// Проверяем корректность MP3, создавая декодер
		_, err := mp3.NewDecoder(bytes.NewReader(audioData))
		if err != nil {
			myLogger.Error("Не удалось создать MP3 декодер", slog.Int64("id", id), slog.Any("error", err))
			return nil, err
		}

		// Копируем аудиоданные в буфер
		_, err = io.Copy(&combinedAudio, bytes.NewReader(audioData))
		if err != nil {
			myLogger.Error("Ошибка объединения аудио", slog.Int64("id", id), slog.Any("error", err))
			return nil, err
		}
		myLogger.Info("Аудио добавлено в объединённый буфер", slog.Int64("id", id))
	}
	myLogger.Info("Успешно объединили все аудиофайлы")

	var RequestForTGAPP *model_kafka.KafkaRequestForTGAPP

	audioDataRequest := combinedAudio.Bytes()

	//region Сохранение аудиодорожки в файл
	err = ioutil.WriteFile("audio.mp3", audioDataRequest, 0644)
	if err != nil {
		myLogger.Error(fmt.Sprintf("Не удалось сохранить аудио", err))
	}
	myLogger.Info("Аудио файл сохранен!")
	//endregion

	RequestForTGAPP = model_kafka.InitKafkaRequestForTTS(messageFromTGAPP.ChatID, audioDataRequest)

	return RequestForTGAPP, nil
}
