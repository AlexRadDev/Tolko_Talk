// Логика работы с текстом сообщений

package service_parser

import (
	"context"
	"fmt"
	"regexp"
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

	// Получаем постов из ТГ канала
	messages, err := s.parser.PostParser(ctx, nameChannel, timePeriod)
	if err != nil {
		mylogger.Error(fmt.Sprintf("Ошибка из PostParser: %v", err))
		return nil, err
	}
	mylogger.Info("Успешно спарсили посты")

	// Создаем карту для хранения обработанных сообщений
	messageMap := make(map[int64]string, 0)

	// Обрабатываем каждое сообщение
	for _, msg := range messages {
		// Очищаем текст сообщения
		cleanedText := FormatText(msg.Text)
		if cleanedText != "" {
			// Используем временную метку как ключ
			messageMap[msg.Timestamp.Unix()] = cleanedText
		}
	}

	return messageMap, nil
}

// FormatText очищает и форматирует текст постов
func FormatText(input string) string {
	// Удаление временных меток и дат через регулярки
	timePatterns := []*regexp.Regexp{
		// regexp.MustCompile(`\b\d{1,2}:\d{2}(:\d{2})?\b`),               // 12:34 или 14:20:59
		regexp.MustCompile(`\b\d{1,2}[./-]\d{1,2}[./-]\d{2,4}\b`),      // 09.05.2025, 9-5-25
		regexp.MustCompile(`\b\d{4}-\d{2}-\d{2} \d{2}:\d{2}(:\d{2})?`), // 2025-05-24 12:34
	}
	for _, re := range timePatterns {
		input = re.ReplaceAllString(input, "")
	}

	// Удаление пустых строк
	lines := strings.Split(input, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	text := strings.Join(nonEmptyLines, "\n")

	// Фильтрация разрешённых символов
	var cleanedText strings.Builder
	for _, char := range text {
		if (char >= 'а' && char <= 'я') ||
			(char >= 'А' && char <= 'Я') ||
			char == 'ё' || char == 'Ё' ||
			(char >= '0' && char <= '9') ||
			char == ' ' || char == '\n' ||
			strings.ContainsRune(".,!?;:—–-()«»\"'", char) {
			cleanedText.WriteRune(char)
		}
	}

	result := cleanedText.String()

	// Обрезка до 4800 символов, безопасно по UTF-8
	if len(result) > 4800 {
		result = result[:4800]
		for len(result) > 0 && !utf8.RuneStart(result[len(result)-1]) {
			result = result[:len(result)-1]
		}
	}

	return result
}
