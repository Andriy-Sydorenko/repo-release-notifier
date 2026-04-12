package main

import (
	"fmt"
	"log"

	"github.com/Andriy-Sydorenko/repo-release-notifier/internal"
	"github.com/Andriy-Sydorenko/repo-release-notifier/internal/database"
)

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPostgres(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&internal.Subscription{}, &internal.ConfirmationToken{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	router := internal.SetupRouter(db, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
