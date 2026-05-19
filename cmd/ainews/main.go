package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"intelhorizons/internal/api"
	"intelhorizons/internal/cache"
	"intelhorizons/internal/store"
	"intelhorizons/internal/worker"
)

const (
	defaultPort         = "8080"
	defaultPollInterval = 5 * time.Minute
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	mode := env("MODE", "both")
	port := env("PORT", defaultPort)
	dbURL := requireEnv("DATABASE_URL")
	redisURL := env("REDIS_URL", "redis://localhost:6379")
	twitterToken := env("TWITTER_BEARER_TOKEN", "")

	st, err := store.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	if err := runMigrations(ctx, st); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	ca, err := cache.New(redisURL)
	if err != nil {
		log.Fatalf("cache: %v", err)
	}
	defer func() {
		if err := ca.Close(); err != nil {
			log.Printf("cache close: %v", err)
		}
	}()

	if err := ca.Ping(ctx); err != nil {
		log.Printf("warn: redis unavailable (%v) — caching disabled", err)
	}

	if mode == "worker" || mode == "both" {
		w := worker.New(st, twitterToken)
		if err := w.Validate(); err != nil {
			log.Fatalf("worker validate: %v", err)
		}
		go w.RunForever(ctx, defaultPollInterval)
	}

	if mode == "server" || mode == "both" {
		h := api.New(st, ca)
		srv := &http.Server{
			Addr:         ":" + port,
			Handler:      h.Router(),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		go func() {
			log.Printf("server: listening on :%s", port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server: %v", err)
			}
		}()

		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	} else {
		<-ctx.Done()
	}

	log.Println("goodbye")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func runMigrations(ctx context.Context, st *store.Store) error {
	pattern := filepath.Join("migrations", "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	migrations := make([]store.Migration, 0, len(files))
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		migrations = append(migrations, store.Migration{
			Filename: filepath.Base(f),
			SQL:      string(b),
		})
	}
	log.Printf("migrations: found %d file(s)", len(migrations))
	return st.RunMigrations(ctx, migrations)
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "fatal: environment variable %s is required\n", key)
		os.Exit(1)
	}
	return v
}
