package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tg_bot/internal/config"
	"tg_bot/tools/logger"
)

// Server представляет HTTP-сервер с поддержкой graceful shutdown
type Server struct {
	srv      *http.Server
	myLogger *slog.Logger
}

// NewServer создаёт новый экземпляр сервера
func NewServer(cfg *config.Config, handler http.Handler) *Server {
	const lblServer = "internal/server/server.go/NewServer()"
	myLogger := logger.NewColorLogger(lblServer)

	// Настраиваем HTTP-сервер
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: handler,
	}

	return &Server{
		srv:      srv,
		myLogger: myLogger,
	}
}

// Start запускает HTTP-сервер в отдельной горутине
func (s *Server) Start() error {
	s.myLogger.Info(fmt.Sprintf("Запуск HTTP-сервера на порту: %v", s.srv.Addr))
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.myLogger.Error(fmt.Sprintf("Ошибка работы сервера: %v", "error", err))
		return err
	}
	return nil
}

// Shutdown выполняет graceful shutdown сервера
func (s *Server) Shutdown(ctx context.Context) error {
	s.myLogger.Info("Инициируем graceful shutdown сервера")
	if err := s.srv.Shutdown(ctx); err != nil {
		s.myLogger.Error(fmt.Sprintf("Ошибка при завершении работы сервера: %v", err))
		return err
	}
	s.myLogger.Info("Сервер успешно завершил работу")
	return nil
}

// WaitForShutdown ожидает сигнал завершения и выполняет graceful shutdown
func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.myLogger.Info("Получен сигнал завершения")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		os.Exit(1)
	}
}
