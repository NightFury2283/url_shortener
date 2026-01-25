package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"url-shortener/internal/config"
)

func main() {
	//TODO: Init Config
	_ = godotenv.Load("../../local.env")

	cfg := config.MustLoad()
	fmt.Println(cfg)

	//TODO: Init Logger log/slog

	//TODO: Init storage sqlite

	//TODO: Init Router

	//TODO: Start Server
}
