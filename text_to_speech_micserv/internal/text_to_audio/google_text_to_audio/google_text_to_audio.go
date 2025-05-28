package google_text_to_audio

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"google.golang.org/api/option"
	texttospeech "google.golang.org/api/texttospeech/v1"

	"text_to_speech_app/internal/config"
	"text_to_speech_app/internal/metrics/metrics_prometheus"
	"text_to_speech_app/internal/model/model_kafka"
	"text_to_speech_app/tools/logger"
)

type GoogleTextToAudio struct {
	clientTTS         *texttospeech.Service
	MetricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus
}

func NewGoogleTextToAudio(cfg *config.Config, metricsPrometheus *metrics_prometheus.TGBotMetricsPrometheus) *GoogleTextToAudio {
	const lblNewTTS = "[myMicro] text_to_speech_micserv/internal/app_text_to_speech/NewGoogleTextToAudio()"
	myLogger := logger.NewColorLogger(lblNewTTS)

	ctx := context.Background()
	clientTTS, err := texttospeech.NewService(ctx, option.WithCredentialsFile(cfg.GoogleCredentialsFile))
	if err != nil {
		myLogger.Error(fmt.Sprintf("[myMicro] Ошибка при создании клиента: %v", err))
		log.Fatal(err)
	}

	return &GoogleTextToAudio{
		MetricsPrometheus: metricsPrometheus,
		clientTTS:         clientTTS,
	}
}

func (s *GoogleTextToAudio) Synthesize(ctx context.Context, messageFromTGAPP *model_kafka.KafkaMessageFromTGAPP) (*model_kafka.KafkaRequestForTGBot, error) {
	const lblSynthesize = "[myMicro] text_to_speech_micserv/internal/app_text_to_speech/Synthesize()"
	myLogger := logger.NewColorLogger(lblSynthesize)

	ctx = context.Background()

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
	myLogger.Info(fmt.Sprintf("Отсортированы ключи мапы: %v", len(keys)))

	var count int64 = 0

	// Итерируем по отсортированным ключам
	for _, id := range keys {
		count++

		text := messageFromTGAPP.TextForAudio[id]

		input := &texttospeech.SynthesisInput{
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

		s.MetricsPrometheus.IncUpdate()
		s.MetricsPrometheus.TimeStartRequestGoogle = time.Now()

		// Выполняем запрос синтеза речи
		resp, err := s.clientTTS.Text.Synthesize(ttsReq).Do()
		if err != nil {
			myLogger.Error("Не удалось синтезировать речь", slog.Int64("id", id), slog.Any("error", err))
			return nil, err
		}
		myLogger.Info(fmt.Sprintf("Успешно синтэзировали речь %v", count))

		tt := time.Since(s.MetricsPrometheus.TimeStartRequestGoogle)
		strChatID := strconv.FormatInt(s.MetricsPrometheus.ChatID, 10)
		strCount := strconv.FormatInt(count, 10)

		myLogger.Info(fmt.Sprintf("Метрики для сохранения: RequestID: %v, ChatID: %v, Time %v, Count %v", s.MetricsPrometheus.RequestID, s.MetricsPrometheus.ChatID, tt, count))
		s.MetricsPrometheus.ObserveDurationGoogle(s.MetricsPrometheus.RequestID, strChatID, strCount, tt)

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

	var RequestForTGAPP *model_kafka.KafkaRequestForTGBot

	audioDataRequest := combinedAudio.Bytes()

	//region сохранение аудио в аудио файл
	//err = s.SaveAudioToFile(audioDataRequest)
	//if err != nil {
	//	myLogger.Error(fmt.Sprintf("Не удалось сохранить аудио в файл", err))
	//}
	//myLogger.Info("Аудио файл сохранен!")
	//endregion

	RequestForTGAPP = model_kafka.InitKafkaRequestForTTS(messageFromTGAPP, audioDataRequest)

	return RequestForTGAPP, nil
}

func (s *GoogleTextToAudio) SaveAudioToFile(audioDataRequest []byte) error {
	err := ioutil.WriteFile("audio.mp3", audioDataRequest, 0644)
	if err != nil {
		return err
	}
	return nil
}
