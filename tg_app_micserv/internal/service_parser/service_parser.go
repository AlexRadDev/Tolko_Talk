// Логика работы с текстом сообщений

package service_parser

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"tg_app_micserv/internal/model/interfaces"
	"tg_app_micserv/tools/logger"
)

// ServiceParser организует получение и очистку сообщений
type ServiceParser struct {
	parser     interfaces.ServiceParser // Интерфейс для получения сообщений
	formatText interfaces.TextCleaner   // Интерфейс для очистки текста
}

// NewMessageService создает новый ServiceParser
func NewServiceParser(parser interfaces.ServiceParser) *ServiceParser {
	return &ServiceParser{
		parser: parser,
	}
}

// FetchMessages парсит и обрабатывает посты(текст) из канала
func (s *ServiceParser) PostParser(ctx context.Context, nameChannel string, timePeriod time.Duration) (string, error) {
	const lbl = "tg_app_micserv/cmd/main.go/main()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	// Получаем сообщения из fetcher (например, Telegram-клиента)
	messages, err := s.parser.PostParser(ctx, nameChannel, timePeriod)
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка из PostParser: %v", err))
		return "", err
	}
	slog.Info("Успешно спарсили посты")

	// TODO: Переделать, что бы каждое сообщение записывалось в мапу, а мапа отсылалась в Kafka
	// Создаем строковый билдер для объединения текста всех сообщений
	var builder strings.Builder
	for _, msg := range messages {
		builder.WriteString(msg.Text + "\n") // Добавляем каждое сообщение с переводом строки
	}

	// Очищаем объединённый текст с помощью функции CleanText
	textForAudio := FormatText(builder.String())
	return textForAudio, nil
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
