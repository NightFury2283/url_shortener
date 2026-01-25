package main

import (
	"fmt"
	"log/slog"
	"os"
	"url-shortener/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	//TODO: Init Config
	_ = godotenv.Load("../../local.env")

	cfg := config.MustLoad()
	fmt.Println(cfg)

	log := setupLogger(cfg.Env)

	log.Info("Logger initialized", slog.String("env", cfg.Env))
	log.Debug("logger debug")

	//TODO: Init Logger log/slog

	//TODO: Init storage sqlite

	//TODO: Init Router

	//TODO: Start Server
}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
		case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
