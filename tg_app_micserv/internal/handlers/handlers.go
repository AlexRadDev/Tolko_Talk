// HTTP обработчики (ручки)

package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"tg_app_micserv/internal/service_parser"
	"tg_app_micserv/tools/logger"
)

// MessageHandler обрабатывает HTTP-запросы, связанные с сообщениями
type MessageHandler struct {
	service *service_parser.ServiceParser
}

// NewMessageHandler создает новый MessageHandler
func NewMessageHandler(service *service_parser.ServiceParser) *MessageHandler {
	return &MessageHandler{ // Возвращает структуру с внедрённым сервисом парсинга
		service: service,
	}
}

// HandlerPostParser обрабатывает конечную точку fetch-messages
func (h *MessageHandler) HandlerPostParser(w http.ResponseWriter, r *http.Request) {
	const lbl = "tg_app_micserv/cmd/main.go/main()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)

	// Получаем значение параметра "channel" из строки запроса
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, `{"error":"channel parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Получаем значение параметра период времени из строки запроса
	hoursStr := r.URL.Query().Get("hours")
	if hoursStr == "" {
		http.Error(w, `{"error":"hours parameter is required"}`, http.StatusBadRequest)
		return
	}
	// Преобразуем параметр "hours" из строки в число с плавающей точкой
	hours, err := strconv.ParseFloat(hoursStr, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid hours parameter"}`, http.StatusBadRequest)
		return
	}

	// Вызываем метод сервиса для получения постов
	messages, err := h.service.PostParser(context.Background(), channel, time.Duration(hours)*time.Hour)
	if err != nil {
		response := APIResponse{Error: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Формируем успешный JSON-ответ с полученными сообщениями
	response := APIResponse{Messages: messages}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// APIResponse представляет структуру ответа JSON
type APIResponse struct {
	Messages string `json:"messages"`
	Error    string `json:"error,omitempty"`
}
