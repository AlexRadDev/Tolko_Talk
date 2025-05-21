// Файл logger.go настраивает логгер на основе библиотеки log/slog с цветным
// выводом для удобства чтения. Соответствует инфраструктурному слою чистой
// архитектуры, обеспечивая единообразное логирование.

package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

// ANSI-коды цветов
const (
	colorReset   = "\033[0m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorBlue    = "\033[36m"
	colorMagenta = "\033[35m"
)

// ColorLevelHandler — обёртка над slog.Handler, которая раскрашивает уровень логирования
type ColorLevelHandler struct {
	handler slog.Handler
	lbl     string // Добавляем поле для префикса lbl
}

// NewColorLogger создаёт slog.Logger с цветной обёрткой
func NewColorLogger(lbl string) *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	// Используем стандартный TextHandler с нашей обёрткой
	baseHandler := slog.NewTextHandler(os.Stdout, opts)
	colorHandler := &ColorLevelHandler{handler: baseHandler, lbl: lbl}
	return slog.New(colorHandler)
}

// Enabled просто передаёт вызов во внутренний handler
func (h *ColorLevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.handler.Enabled(context.Background(), level)
}

func (h *ColorLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColorLevelHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *ColorLevelHandler) WithGroup(name string) slog.Handler {
	return &ColorLevelHandler{handler: h.handler.WithGroup(name)}
}

func (h *ColorLevelHandler) Handle(ctx context.Context, record slog.Record) error {
	level := record.Level.String()
	var coloredLevel string

	switch record.Level {
	case slog.LevelInfo:
		coloredLevel = colorGreen + level + colorReset
	case slog.LevelWarn:
		coloredLevel = colorYellow + level + colorReset
	case slog.LevelError:
		coloredLevel = colorRed + level + colorReset
	case slog.LevelDebug:
		coloredLevel = colorBlue + level + colorReset
	default:
		coloredLevel = level
	}

	// Добавляем префикс lbl с цветом colorMagenta перед сообщением
	prefixedMessage := colorMagenta + h.lbl + colorReset + " " + record.Message

	// Выводим кастомный лог в stdout
	writer := &strings.Builder{}
	writer.WriteString(time.Now().Format("2006/01/02 15:04:05 "))
	writer.WriteString(coloredLevel + " ") // "[" + coloredLevel + "] "
	writer.WriteString(prefixedMessage + "\n")

	_, err := os.Stdout.Write([]byte(writer.String()))
	if err != nil {
		return err
	}

	// Также передаём оригинальную запись во вложенный handler (комментируем, если не нужно)
	return nil
}
