package model_kafka

// KafkaMessageFromTGB представляет структуру данных сообщения из topic Kafka ("tgbot-tgapp")
type KafkaMessageFromTGB struct {
	ChatID               int64   `json:"ChatID"`               // идентификатор чата Telegram
	NameChanel           string  `json:"NameChanel"`           // имя или ссылка на Telegram-канал
	SpeakingRate         float64 `json:"SpeakingRate"`         // скорость речи
	TimePeriod           int     `json:"TimePeriod"`           // период времени в часах
	AwaitingChannelInput bool    `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
}

// KafkaRequestForTTS представляет структуру данных для сообщения в topic Kafka "tgapp-texttospeech"
type KafkaRequestForTTS struct {
	ChatID               int64            `json:"ChatID"`               // идентификатор чата Telegram
	SpeakingRate         float64          `json:"SpeakingRate"`         // скорость речи
	AwaitingChannelInput bool             `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
	TextForAudio         map[int64]string `json:"TextForAudio"`         // Мапа с постами, для отправки по очереди в text to speech Google
}

func InitKafkaRequestForTTS(message *KafkaMessageFromTGB, textForAudio map[int64]string) *KafkaRequestForTTS {
	return &KafkaRequestForTTS{
		ChatID:               message.ChatID,
		SpeakingRate:         message.SpeakingRate,
		AwaitingChannelInput: message.AwaitingChannelInput,
		TextForAudio:         textForAudio,
	}
}

// KafkaRequestForTGBot представляет структуру данных для сообщения в topic Kafka "tgapp-texttospeech"
type KafkaRequestForTGBot struct {
	ChatID               int64  `json:"ChatID"`               // идентификатор чата Telegram
	AwaitingChannelInput bool   `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
	TextForUsers         string `json:"TextForUsers"`         // Текст что постов за это время нет
}
