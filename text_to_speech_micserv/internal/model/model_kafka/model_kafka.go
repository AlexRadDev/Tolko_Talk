package model_kafka

// KafkaMessageFromTGB представляет структуру данных сообщения из topic Kafka ("tgbot-tgapp")
type KafkaMessageFromTGAPP struct {
	ChatID               int64            `json:"ChatID"`               // идентификатор чата Telegram
	SpeakingRate         float64          `json:"SpeakingRate"`         // скорость речи
	AwaitingChannelInput bool             `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
	TextForAudio         map[int64]string `json:"TextForAudio"`         // Мапа с постами, для отправки по очереди в text to speech Google
}

// KafkaRequestForTGAPP представляет структуру данных для сообщения в topic Kafka "tgapp-texttospeech"
type KafkaRequestForTGAPP struct {
	ChatID    int64  `json:"ChatID"`     // идентификатор чата Telegram
	AudioData []byte `json:"audio_data"` // аудиоданные в формате MP3
}

func InitKafkaRequestForTTS(chatID int64, combinedAudio []byte) *KafkaRequestForTGAPP {
	return &KafkaRequestForTGAPP{
		ChatID:    chatID,
		AudioData: combinedAudio,
	}
}
