package model_text_to_speech

// TextToSpeechRequest представляет запрос на преобразование текста в речь
type TextToSpeechRequest struct {
	Text         map[int64]string `json:"text"`          // текст для синтеза речи
	SpeakingRate float64          `json:"speaking_rate"` // скорость речи (например, 1.0 — стандартная)
}

// TextToSpeechResponse представляет ответ с синтезированным аудио
type TextToSpeechResponse struct {
	AudioData []byte `json:"audio_data"` // аудиоданные в формате MP3
	Error     string `json:"error,omitempty"`
}
