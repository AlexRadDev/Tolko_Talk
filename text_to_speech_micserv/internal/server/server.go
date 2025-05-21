// Файл server.go реализует HTTP-сервер микросервиса Text-to-Speech.
// Отвечает за обработку входящих HTTP-запросов, десериализацию данных и вызов бизнес-логики.

package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"text_to_speech_app/internal/app_text_to_speech"
	"text_to_speech_app/internal/model/model_text_to_speech"
	"text_to_speech_app/tools/logger"
)

// Server представляет HTTP-сервер
type Server struct {
	srv        *http.Server                // объект HTTP-сервера
	ttsService *app_text_to_speech.Service // указатель на сервис Text-to-Speech
}

// NewServer создаёт новый HTTP-сервер
func NewServer(port string, ttsService *app_text_to_speech.Service) *Server {
	// Создаём мультиплексор для маршрутизации
	mux := http.NewServeMux()
	srv := &Server{
		srv: &http.Server{
			Addr:    port, // Устанавливаем порт
			Handler: mux,  // Устанавливаем мультиплексор
		},
		ttsService: ttsService, // Внедряем сервис
	}
	mux.HandleFunc("/synthesize", srv.handleSynthesize)

	return srv
}

// handleSynthesize обрабатывает POST-запросы на /synthesize
func (s *Server) handleSynthesize(w http.ResponseWriter, r *http.Request) {
	const loghandleSynthesize = "internal/infrastructure/server/server.go"
	myLogger := logger.NewColorLogger(loghandleSynthesize)

	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		myLogger.Error("Метод не поддерживается", slog.String("method", r.Method))
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Создаём структуру для запроса
	var req model_text_to_speech.TextToSpeechRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		myLogger.Error("Ошибка декодирования JSON", slog.Any("error", err))
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	myLogger.Info("Успешно декодировали JSON", slog.Any("request", req))

	// Вызываем бизнес-логику для синтеза речи
	resp, err := s.ttsService.Synthesize(context.Background(), &req)
	if err != nil {
		myLogger.Error("Ошибка синтеза речи", slog.Any("error", err))
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Проверяем, есть ли ошибка в ответе
	if resp.Error != "" {
		myLogger.Error("Ошибка в ответе сервиса", slog.String("error", resp.Error))
		http.Error(w, resp.Error, http.StatusBadRequest)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		myLogger.Error("Ошибка кодирования JSON", slog.Any("error", err))
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	myLogger.Info("Успешно отправили ответ клиенту")
}

// ListenAndServe запускает HTTP-сервер
func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

// Shutdown выполняет graceful shutdown сервера
func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
