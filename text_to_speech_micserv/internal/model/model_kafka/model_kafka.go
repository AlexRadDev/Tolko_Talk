package model_kafka

// KafkaMessageFromTGB представляет структуру данных сообщения из topic Kafka ("tgapp-texttospeech")
type KafkaMessageFromTGAPP struct {
	ChatID       int64            `json:"ChatID"`
	RequestID    string           `json:"RequestID"`
	SpeakingRate float64          `json:"SpeakingRate"`
	TextForAudio map[int64]string `json:"TextForAudio"`
}

// KafkaRequestForTGBot представляет структуру данных для сообщения в topic Kafka "texttospeech-tgbot"
type KafkaRequestForTGBot struct {
	ChatID    int64  `json:"ChatID"`
	RequestID string `json:"RequestID"`
	AudioData []byte `json:"audio_data"` // аудиоданные в формате MP3
}

func InitKafkaRequestForTTS(MessageFromTGAPP *KafkaMessageFromTGAPP, combinedAudio []byte) *KafkaRequestForTGBot {
	return &KafkaRequestForTGBot{
		ChatID:    MessageFromTGAPP.ChatID,
		RequestID: MessageFromTGAPP.RequestID,
		AudioData: combinedAudio,
	}
}
