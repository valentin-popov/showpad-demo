package main

import (
	"context"
	"gateway/pkg/config"
	limiter "gateway/pkg/limiting-service"
	"log"
	"os/signal"
	"syscall"
)

func main() {

	cfg, err := config.Load("config/gateway.hcl")

	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}
	conf, err := cfg.Parse()
	if err != nil {
		log.Fatal("Failed to parse config: ", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	limiter, err := limiter.New(ctx, conf)

	if err := limiter.Run(ctx); err != nil {
		limiter.Stop()
	}

}
