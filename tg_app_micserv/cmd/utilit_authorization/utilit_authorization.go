package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/joho/godotenv"

	"tg_app_micserv/internal/config"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	//"github.com/gotd/td/telegram/kv"
	"tg_app_micserv/tools/logger"
)

const pathEnv = ".env"

func main() {
	const lbl = "tg_app_micserv/cmd/utilit_authorization/utilit_authorization.go/main()"
	logger := logger.NewColorLogger(lbl)
	slog.SetDefault(logger)
	slog.Info("Запустили обертку logger")

	err := godotenv.Load(pathEnv)
	if err != nil {
		slog.Error("Неудалось найти файл .env", "error", err)
		log.Fatal(err)
	}
	slog.Info("Подгрузили файл .env")

	// Загрузка конфигов
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Не удалось загрузить конфиг", "error", err)
		log.Fatal(err)
	}
	slog.Info("Успешно создали объект конфига")

	storage := &telegram.FileSessionStorage{
		Path: "session.json",
	}

	client := telegram.NewClient(cfg.API_ID, cfg.API_Hash, telegram.Options{
		SessionStorage: storage,
	})

	err := client.Run(context.Background(), func(ctx context.Context) error {
		flow := auth.NewFlow(
			auth.Constant(cfg.Phone, cfg.Two_F_Password, auth.SendCodeOptions{}),
			auth.SendCodeOptions{},
		)
		return flow.Run(ctx, client.Auth())
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Сессия сохранена!")
}
