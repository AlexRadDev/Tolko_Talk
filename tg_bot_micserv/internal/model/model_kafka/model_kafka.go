package model_kafka

// KafkaMessageFromTGAPP представляет структуру сообщения из topic Kafka ("tgbot-tgapp")
type KafkaMessageFromTGAPP struct {
	ChatID               int64            `json:"ChatID"`               // идентификатор чата Telegram
	SpeakingRate         float64          `json:"SpeakingRate"`         // скорость речи
	AwaitingChannelInput bool             `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
	TextForAudio         map[int64]string `json:"TextForAudio"`         // Мапа с постами, для отправки по очереди в text to speech Google
}

// KafkaRequestForTGBOT представляет структуру данных для получения сообщения из topic "texttospeech-tgbot"
type KafkaRequestForTGBOT struct {
	ChatID    int64  `json:"ChatID"`     // идентификатор чата Telegram
	AudioData []byte `json:"audio_data"` // аудиоданные в формате MP3
}

// KafkaRequestForTGappTGbot представляет структуру сообщения для topic "tgapp-tgbot"
type KafkaRequestForTGappTGbot struct {
	ChatID               int64  `json:"ChatID"`               // идентификатор чата Telegram
	AwaitingChannelInput bool   `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
	TextForUsers         string `json:"TextForUsers"`         // Текст что постов за это время нет
}
