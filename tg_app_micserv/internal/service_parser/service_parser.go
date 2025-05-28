package service_parser

import (
	"regexp"
	"strings"
	"unicode/utf8"

	tg_post_model "tg_app_micserv/internal/model"
)

func FormationTextForTTS(messages []tg_post_model.Message) (map[int64]string, error) {
	messageMap := make(map[int64]string, 0)
	for _, msg := range messages {
		cleanedText := FormatText(msg.Text) // Очищаем текст сообщения
		if cleanedText != "" {
			messageMap[msg.Timestamp.Unix()] = cleanedText // Используем временную метку как ключ
		}
	}

	return messageMap, nil
}

// FormatText очищает и форматирует текст постов
func FormatText(input string) string {
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

// ExtractChannelName Форматируем имя канала, превращает ссылку в нужный формат для TG API PARSER
func ExtractChannelName(input string) string {
	if strings.HasPrefix(input, "@") {
		return input[1:]
	}
	return strings.TrimPrefix(input, "https://t.me/")
}
