package app_text_to_speech

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"google.golang.org/api/option"
	texttospeech "google.golang.org/api/texttospeech/v1"
)

// SynthesizeText по входящему тексту синтэзирует речь
func SynthesizeText(text, keyToSpeech, outputFile string, speakingRate float64) error {
	// Создать контекст
	ctx := context.Background()

	// Создайте новый клиент преобразования текста в речь
	client, err := texttospeech.NewService(ctx, option.WithCredentialsFile(keyToSpeech))
	if err != nil {
		return fmt.Errorf("failed to create Text-to-Speech client: %v", err)
	}

	// Определите входной текст
	input := &texttospeech.SynthesisInput{
		Text: text,
	}

	// Настройте голос (например, код языка «en-US» и женский голос)
	//voice := &texttospeech.VoiceSelectionParams{
	//	LanguageCode: "en-US",
	//	Name:         "en-US-Wavenet-D",
	//	SsmlGender:   "FEMALE",
	//}

	voice := &texttospeech.VoiceSelectionParams{
		LanguageCode: "ru-RU",           // Код языка для русского
		Name:         "ru-RU-Wavenet-A", // Имя голоса (WaveNet голос для русского языка)
		SsmlGender:   "MALE",            // Пол голоса
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
		return fmt.Errorf("failed to synthesize speech: %v", err)
	}

	// Декодируем base64 строку в байты
	audioData, err := base64.StdEncoding.DecodeString(resp.AudioContent)
	if err != nil {
		return fmt.Errorf("failed to decode base64 audio content: %v", err)
	}

	// Записать аудиоконтент в файл
	err = ioutil.WriteFile(outputFile, audioData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write audio file: %v", err)
	}

	fmt.Printf("Audio content written to %s\n", outputFile)
	return nil
}
