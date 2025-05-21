package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"tg_app_micserv/internal/handlers"
	"tg_app_micserv/tools/logger"
)

// Server инкапсулирует HTTP-сервер и его конфигурацию
type Server struct {
	srv *http.Server
}

// NewServer создает новый экземпляр Server
func NewServer(port string, handler *handlers.MessageHandler) *Server {
	// Настройка HTTP-сервера
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      nil, // Используем дефолтный http.DefaultServeMux
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Регистрация обработчика эндпоинта
	http.HandleFunc("/post_parser", handler.HandlerPostParser)

	return &Server{
		srv: srv,
	}
}

// Run запускает HTTP-сервер
func (s *Server) Run() error {
	const lbl = "tg_app_micserv/internal/server/server.go/Run()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)
	slog.Info(fmt.Sprintf("Старт сервера на порту %v", s.srv.Addr))

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown выполняет мягкую остановку сервера
func (s *Server) Shutdown(ctx context.Context) error {
	//s.logger.Info("Инициируется остановка сервера")
	if err := s.srv.Shutdown(ctx); err != nil {
		//s.logger.Error("Ошибка при остановке сервера", "error", err)
		return err
	}
	//s.logger.Info("Сервер успешно остановлен")
	return nil
}
