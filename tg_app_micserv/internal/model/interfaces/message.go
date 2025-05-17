// Доменные интерфейсы

package interfaces

import (
	"context"
	"time"

	"tg_app_micserv/internal/model"
)

// ServiceParser интерфейс для получения сообщений
type ServiceParser interface {
	PostParser(ctx context.Context, channel string, period time.Duration) ([]tg_post_model.Message, error)
}

// TextCleaner интерфейс для очистки текста сообщения
type TextCleaner interface {
	FormatText(text string) string
}
