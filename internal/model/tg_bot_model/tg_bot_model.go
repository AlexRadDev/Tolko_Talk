package tg_bot_model

type TgBotRequest struct { // запрос
	NameChanel   string  `json:"name_chanel"`
	TimePeriod   int     `json:"time_period"`
	SpeakingRate float64 `json:"speaking_rate"`
}
