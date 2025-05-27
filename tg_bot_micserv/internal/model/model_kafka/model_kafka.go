package model_kafka

// KafkaMessageFromTGAPP представляет структуру сообщения для topic Kafka ("tgbot-tgapp")
type KafkaMessageFromTGAPP struct {
	ChatID               int64   `json:"ChatID"`
	RequestID            string  `json:"RequestID"`
	NameChanel           string  `json:"NameChanel"`
	SpeakingRate         float64 `json:"SpeakingRate"`
	TimePeriod           int     `json:"TimePeriod"`
	AwaitingChannelInput bool    `json:"AwaitingChannelInput"` // флаг, указывающий, ожидается ли ввод имени канала
}

// KafkaAnswerFromTTS представляет структуру данных для получения ответа из topic "texttospeech-tgbot"
type KafkaAnswerFromTTS struct {
	ChatID    int64  `json:"ChatID"`
	RequestID string `json:"RequestID"`
	AudioData []byte `json:"audio_data"`
}

// KafkaRequestForTGappTGbot представляет структуру сообщения из topic "tgapp-tgbot"
type KafkaRequestForTGappTGbot struct {
	ChatID               int64  `json:"ChatID"`
	RequestID            string `json:"RequestID"`
	AwaitingChannelInput bool   `json:"AwaitingChannelInput"`
	TextForUsers         string `json:"TextForUsers"`
}
