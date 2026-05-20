package main

import (
	"log"
	"sports-platform/internal/config"
	"sports-platform/internal/repository"
	"sports-platform/internal/router"
	"sports-platform/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	// Seed initial super admin from .env
	adminRepo := repository.NewAdminRepo(db)
	adminSvc := service.NewAdminService(adminRepo, cfg)
	adminSvc.SeedSuperAdmin()

	r := router.Setup(cfg, db)

	log.Printf("Server starting on :%s (env: %s)", cfg.Port, cfg.AppEnv)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
