package bot_request

// Структура TgBotRequest представляет запрос пользователя(клиента) к боту
type TgBotRequest struct {
	ChatID               int64   // идентификатор чата Telegram
	NameChanel           string  // имя или ссылка на Telegram-канал
	SpeakingRate         float64 // скорость речи
	TimePeriod           int     // период времени в часах
	AwaitingChannelInput bool    // флаг, указывающий, ожидается ли ввод имени канала
}
