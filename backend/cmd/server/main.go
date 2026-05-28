package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"trbb/internal/api"
	"trbb/internal/admin"
	"trbb/internal/third"
	"trbb/internal/config"
	"trbb/internal/middleware"
	"trbb/pkg/database"
	"trbb/pkg/cache"
	"trbb/pkg/logger"
	"trbb/pkg/storage"
)

func main() {
	// ── Config ──────────────────────────────────────────────
	cfg := config.Load()

	// ── Logger ──────────────────────────────────────────────
	log := logger.New(cfg.Log)
	log.Info("TRBB backend starting...")

	// ── Database ────────────────────────────────────────────
	db, err := database.New(cfg.DB)
	if err != nil {
		log.Fatal("failed to connect database", "error", err)
	}
	defer db.Close()

	// ── Redis ───────────────────────────────────────────────
	rdb, err := cache.New(cfg.Redis)
	if err != nil {
		log.Fatal("failed to connect redis", "error", err)
	}
	defer rdb.Close()

	// ── MinIO ───────────────────────────────────────────────
	minio, err := storage.New(cfg.Minio)
	if err != nil {
		log.Fatal("failed to connect minio", "error", err)
	}

	// ── Gin ─────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS())
	r.Use(middleware.RealIP())

	// ── Routes ──────────────────────────────────────────────
	v1 := r.Group("/v1")
	api.RegisterRoutes(v1.Group("/api"), db, rdb, minio, cfg, log)
	admin.RegisterRoutes(v1.Group("/admin"), db, rdb, minio, cfg, log)
	third.RegisterRoutes(v1.Group("/third"), db, minio, cfg, log)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "trbb-backend"})
	})

	// ── Server ──────────────────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", "error", err)
		}
	}()

	// ── Graceful Shutdown ────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", "error", err)
	}
	log.Info("server exited")
}
