package main

import (
	"api/pkg/config"
	"api/pkg/server"
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {

	cfg, err := config.Load("config/api.hcl")
	// cfg, err := config.Load("/Users/valentin/Documents/dev/showpad-demo/api/config/api.hcl")

	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}
	conf, err := cfg.Parse()
	if err != nil {
		log.Fatal("Failed to parse config: ", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv := server.New(conf)

	if err := srv.Run(ctx); err != nil {
		log.Fatal("Failed to parse config: ", err)
	}

}
