package model_kafka

// KafkaMessageFromTGB представляет структуру данных сообщения из topic Kafka ("tgbot-tgapp")
type KafkaMessageFromTGB struct {
	ChatID       int64   `json:"ChatID"`
	RequestID    string  `json:"RequestID"`
	NameChanel   string  `json:"NameChanel"`
	SpeakingRate float64 `json:"SpeakingRate"`
	TimePeriod   int     `json:"TimePeriod"`
}

// KafkaRequestForTTS представляет структуру данных для сообщения в topic Kafka "tgapp-texttospeech"
type KafkaRequestForTTS struct {
	ChatID       int64            `json:"ChatID"`
	RequestID    string           `json:"RequestID"`
	SpeakingRate float64          `json:"SpeakingRate"`
	TextForAudio map[int64]string `json:"TextForAudio"`
}

// InitKafkaRequestForTTS конструктор запроса в TTS микросервис
func InitKafkaRequestForTTS(message *KafkaMessageFromTGB, textForAudio map[int64]string) *KafkaRequestForTTS {
	return &KafkaRequestForTTS{
		ChatID:       message.ChatID,
		RequestID:    message.RequestID,
		SpeakingRate: message.SpeakingRate,
		TextForAudio: textForAudio,
	}
}

// KafkaRequestForTGBot представляет структуру данных для сообщения в topic Kafka "tgapp-texttospeech"
type KafkaRequestForTGBot struct {
	ChatID               int64  `json:"ChatID"`
	RequestID            string `json:"RequestID"`
	AwaitingChannelInput bool   `json:"AwaitingChannelInput"`
	TextForUsers         string `json:"TextForUsers"`
}

func InitKafkaRequestForTGBot(message *KafkaMessageFromTGB, text string) *KafkaRequestForTGBot {
	return &KafkaRequestForTGBot{
		ChatID:       message.ChatID,
		RequestID:    message.RequestID,
		TextForUsers: text,
	}
}
