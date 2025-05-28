package tg_post_model

import "time"

type Message struct {
	Text      string
	Timestamp time.Time
}
