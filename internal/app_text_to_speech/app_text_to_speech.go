// https://console.cloud.google.com/apis/library/texttospeech.googleapis.com?inv=1&invt=AbxYSA&project=semiotic-mender-459813-c9

package app_text_to_speech

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"tolko_talk/tools/logger"

	"google.golang.org/api/option"
	texttospeech "google.golang.org/api/texttospeech/v1"
)

// SynthesizeText по входящему тексту синтэзирует речь
func SynthesizeText(text, keyToSpeech, outputFile string, speakingRate float64) ([]byte, error) {
	// Метка функции для логов
	const lbl = "internal/app_text_to_speech/app_text_to_speech.go/SynthesizeText()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	// Создать контекст
	ctx := context.Background()

	// Создайте новый клиент преобразования текста в речь
	client, err := texttospeech.NewService(ctx, option.WithCredentialsFile(keyToSpeech))
	if err != nil {
		return nil, fmt.Errorf("Не удалось создать клиент Text-to-Speech: %v", err)
	}
	slog.Info("Создали клиента Text-to-Speech")

	// Определите входной текст
	input := &texttospeech.SynthesisInput{
		Text: text,
	}
	// Настройте голоса
	voice := &texttospeech.VoiceSelectionParams{
		LanguageCode: "ru-RU",            // Код языка для русского ("en-US",)
		Name:         "ru-RU-Standard-B", // Имя голоса (WaveNet голос для русского языка) ("en-US-Wavenet-D")
		SsmlGender:   "FEMALE",           // Пол голоса (MALE)
	}
	// Настройте параметры звука (например, формат MP3)
	audioConfig := &texttospeech.AudioConfig{
		AudioEncoding: "MP3",
		SpeakingRate:  speakingRate, // скорость произношение, по умолчанию 1.0
	}
	// Создайте запрос для синтеза речи
	req := &texttospeech.SynthesizeSpeechRequest{
		Input:       input,
		Voice:       voice,
		AudioConfig: audioConfig,
	}

	// Выполните запрос синтеза речи
	resp, err := client.Text.Synthesize(req).Do()
	if err != nil {
		return nil, fmt.Errorf("Не удалось синтезировать речь: %v", err)
	}
	slog.Info("Успешно синтэзировали речь")

	// Декодируем base64 строку в байты
	audioData, err := base64.StdEncoding.DecodeString(resp.AudioContent)
	if err != nil {
		return nil, fmt.Errorf("не удалось декодировать аудио base64 в []byte: %v", err)
	}
	slog.Info("Успешно декодировали аудио base64 в []byte")

	//region Сохранение аудиодорожки в файл
	//err = ioutil.WriteFile(outputFile, audioData, 0644)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to write audio file: %v", err)
	//}
	//slog.Info(fmt.Sprintf("Audio content written to %s\n", outputFile))
	//endregion

	return audioData, nil
}
