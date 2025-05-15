package tg_bot_handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tolko_talk/internal/tg_bot_usecase"
)

// Определение структуры Handler, которая будет обрабатывать входящие сообщения
type Handler struct {
	uc *tg_bot_usecase.UseCase // Указатель на слой UseCase, содержащий бизнес-логику бота
}

// NewHandler — конструктор для инициализации Handler с внедрением зависимости UseCase
func NewHandler(uc *tg_bot_usecase.UseCase) *Handler {
	return &Handler{uc: uc}
}

// HandleMessage — основной метод обработки входящих сообщений Telegram
func (h *Handler) HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg == nil {
		return
	}
	// Передаём управление бизнес-логике UseCase
	err := h.uc.HandleMessage(bot, msg.Chat.ID, msg.Text)
	if err != nil {
		// обработка ошибок, логгирование
	}
}

// HandleMessage — основной метод обработки входящих сообщений Telegram
//func (h *Handler) HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
//	// Проверяем, является ли сообщение командой (начинается с /)
//	if msg.IsCommand() {
//		switch msg.Command() { // Обработка команды в зависимости от её имени
//		case "start": // Если команда "start", вызываем соответствующую функцию usecase
//			h.uc.StartCommand(bot, msg)
//		default: // Если команда неизвестна, вызываем обработчик неизвестных команд
//			h.uc.UnknownCommand(bot, msg)
//		}
//	} else {
//		h.uc.EchoMessage(bot, msg) // Если сообщение не является командой — вызываем EchoMessage (можно заменить на парсинг)
//	}
//}

// ParseMessageToRequest разбирает текст сообщения Telegram в структуру TgBotRequest
//func ParseMessageToRequest(message string) (*tg_bot_model.TgBotRequest, error) {
//	req := &tg_bot_model.TgBotRequest{} // Создаём переменную результата
//
//	lines := strings.Split(message, "\n") // Разбиваем текст по строкам
//	for _, line := range lines {
//		parts := strings.SplitN(line, ":", 2) // Отделяем ключ и значение (по символу ':')
//		if len(parts) != 2 {
//			return nil, fmt.Errorf("не удалось разобрать строку: %s", line)
//		}
//
//		// Удаляем лишние пробелы
//		key := strings.TrimSpace(parts[0])
//		value := strings.TrimSpace(parts[1])
//
//		// Сопоставляем ключи с полями структуры
//		switch strings.ToLower(key) {
//		case "канал":
//			req.NameChanel = value
//		case "период":
//			period, err := strconv.Atoi(value)
//			if err != nil {
//				return nil, fmt.Errorf("неверный формат периода: %v", err)
//			}
//			req.TimePeriod = period
//		case "скорость":
//			rate, err := strconv.ParseFloat(value, 64)
//			if err != nil {
//				return nil, fmt.Errorf("неверный формат скорости: %v", err)
//			}
//			req.SpeakingRate = rate
//		default:
//			return nil, fmt.Errorf("неизвестное поле: %s", key)
//		}
//	}
//
//	return req, nil
//}
