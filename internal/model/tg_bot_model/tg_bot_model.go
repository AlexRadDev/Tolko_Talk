package tg_bot_model

import "time"

type TgBotRequest struct { // запрос
	NameChanel           string        `json:"name_chanel"`
	TimePeriod           time.Duration `json:"time_period"`
	SpeakingRate         float64       `json:"speaking_rate"`
	AwaitingChannelInput bool          `json:"awaiting_channel_input"`
}
