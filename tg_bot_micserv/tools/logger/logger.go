package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	colorReset   = "\033[0m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorBlue    = "\033[36m"
	colorMagenta = "\033[35m"
	colorOrange  = "\033[38;5;208m" // Для #ff9d1e
	colorLime    = "\033[38;5;190m" // Для #c8ff32
)

// LevelMetrics — новый уровень логирования
const LevelMetrics slog.Level = 12

// Logger — обёртка над slog.Logger
type Logger struct {
	*slog.Logger
}

// ColorLevelHandler — обёртка над slog.Handler, которая раскрашивает уровень логирования
type ColorLevelHandler struct {
	handler slog.Handler
	lbl     string // Добавляем поле для префикса lbl
}

// NewColorLogger создаёт Logger с цветной обёрткой
func NewColorLogger(lbl string) *Logger {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug, // Включает все уровни, включая Metrics
	}

	baseHandler := slog.NewTextHandler(os.Stdout, opts)

	colorHandler := &ColorLevelHandler{handler: baseHandler, lbl: lbl}

	return &Logger{slog.New(colorHandler)}
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
	var level string
	var coloredLevel string

	// Кастомное строковое представление для уровней
	switch record.Level {
	case slog.LevelDebug:
		level = "DEBUG"
		coloredLevel = colorBlue + level + colorReset
	case slog.LevelInfo:
		level = "INFO"
		coloredLevel = colorGreen + level + colorReset
	case slog.LevelWarn:
		level = "WARN"
		coloredLevel = colorYellow + level + colorReset
	case slog.LevelError:
		level = "ERROR"
		coloredLevel = colorRed + level + colorReset
	case LevelMetrics:
		level = "METRICS"
		coloredLevel = colorLime + level + colorReset
	default:
		level = record.Level.String() // Используем стандартное представление для других уровней
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

	return nil
}

// Metrics — кастомный метод для логирования на уровне LevelMetrics
func (l *Logger) Metrics(msg string, args ...any) {
	l.Log(context.Background(), LevelMetrics, msg, args...)
}
