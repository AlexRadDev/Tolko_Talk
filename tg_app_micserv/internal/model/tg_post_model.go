// Доменные структуры

package tg_post_model

import "time"

// Message структура для постов

type Message struct {
	Text      string
	Timestamp time.Time
}
