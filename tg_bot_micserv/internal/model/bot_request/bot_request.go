// Файл bot_request.go определяет доменную модель TgBotRequest, которая представляет запрос пользователя в Telegram-боте.
// Модель содержит данные о выбранном канале, скорости речи, периоде времени и состоянии ввода.

package bot_request

// Структура TgBotRequest представляет запрос пользователя к боту
type TgBotRequest struct {
	ChatID               int64   // идентификатор чата Telegram
	NameChanel           string  // имя или ссылка на Telegram-канал
	SpeakingRate         float64 // скорость речи
	TimePeriod           int     // период времени в часах
	AwaitingChannelInput bool    // флаг, указывающий, ожидается ли ввод имени канала
}
