package main

import (
	"github.com/shammalie/web-forum/internal/server"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()

	server.Serve(logger)
}
