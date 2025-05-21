// Логика работы с текстом сообщений

package service_parser

import (
	"context"
	"fmt"
	"strings"
	"tg_app_micserv/internal/kafka/producer"
	"time"
	"unicode/utf8"

	"tg_app_micserv/internal/model/interfaces"
	"tg_app_micserv/tools/logger"
)

// ServiceParser организует получение и очистку сообщений
type ServiceParser struct {
	parser     interfaces.ServiceParser // Интерфейс для получения сообщений
	producer   producer.MessageProducer // Интерфейс для отправки в Kafka
	formatText interfaces.TextCleaner   // Интерфейс для очистки текста
}

// NewMessageService создает новый ServiceParser
func NewServiceParser(parser interfaces.ServiceParser) *ServiceParser {
	return &ServiceParser{
		parser: parser,
	}
}

// FetchMessages парсит и обрабатывает посты(текст) из канала
func (s *ServiceParser) PostParser(ctx context.Context, nameChannel string, timePeriod time.Duration) (map[int64]string, error) {
	const lbl = "tg_app_micserv/internal/service_parser/service_parser.go/PostParser()"
	mylogger := logger.NewColorLogger(lbl)

	// Форматируем имя канала
	channelNameForParserFunc := ""
	if len(nameChannel) > 0 && nameChannel[0] == '@' {
		channelNameForParserFunc = nameChannel[1:]
	}
	const prefix = "https://t.me/"
	if len(nameChannel) > 0 && strings.HasPrefix(nameChannel, prefix) {
		channelNameForParserFunc = strings.TrimPrefix(nameChannel, prefix)
	}
	mylogger.Info(fmt.Sprintf("Имя канала успешно извлечено: %v", channelNameForParserFunc))

	// Получаем сообщения из fetcher (например, Telegram-клиента)
	messages, err := s.parser.PostParser(ctx, channelNameForParserFunc, timePeriod)
	if err != nil {
		mylogger.Error(fmt.Sprintf("Ошибка из PostParser: %v", err))
		return nil, err
	}
	mylogger.Info("Успешно спарсили посты")

	// Создаем карту для хранения обработанных сообщений
	messageMap := make(map[int64]string)

	// Обрабатываем каждое сообщение
	for _, msg := range messages {
		// Очищаем текст сообщения
		cleanedText := FormatText(msg.Text)
		if cleanedText != "" {
			// Используем временную метку как ключ
			messageMap[msg.Timestamp.Unix()] = cleanedText
		}
	}

	//mylogger.Info(fmt.Sprintf("Итоговый текст из функции PostParser(): %v", messageMap))
	//// Отправляем карту в Kafka
	//if err := s.producer.ProduceMessages(ctx, messageMap); err != nil {
	//	return "", fmt.Errorf("ошибка отправки в Kafka: %w", err)
	//}

	return messageMap, nil
}

//----------------------------------------------------------------------------------------------------------------------

// FormatText очищает и форматирует текст постов
func FormatText(input string) string {
	lines := strings.Split(input, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	text := strings.Join(nonEmptyLines, "\n")

	var cleanedText strings.Builder
	for _, char := range text {
		if (char >= 'а' && char <= 'я') ||
			(char >= 'А' && char <= 'Я') ||
			char == 'ё' || char == 'Ё' ||
			string(char) == " " ||
			string(char) == "." || string(char) == "," || string(char) == "!" ||
			string(char) == "?" || string(char) == ";" || string(char) == ":" ||
			string(char) == "—" || string(char) == "–" || string(char) == "-" ||
			string(char) == "(" || string(char) == ")" || string(char) == "«" ||
			string(char) == "»" || string(char) == "\"" || string(char) == "'" ||
			char == '\n' {
			cleanedText.WriteRune(char)
		}
	}

	result := cleanedText.String()

	if len(result) > 4800 {
		result = result[:4800]
		for len(result) > 0 && !utf8.RuneStart(result[len(result)-1]) {
			result = result[:len(result)-1]
		}
	}

	return result
}
