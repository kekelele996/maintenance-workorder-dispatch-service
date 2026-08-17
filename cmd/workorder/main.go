package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"workorder/internal/config"
	"workorder/internal/handler"
	"workorder/internal/repository"
	"workorder/internal/service"
	"workorder/internal/store"
	"workorder/internal/worker"
)

func main() {
	cfg := config.Load()
	st := store.New()
	repo := repository.New(st)
	svc := service.New(repo, cfg)
	scheduler := worker.New(repo, svc, time.Duration(cfg.PollIntervalMs)*time.Millisecond)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := scheduler.Run(ctx); err != nil && err != context.Canceled {
			log.Printf("scheduler stopped: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler.New(svc).Routes(),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Println("workorder service listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
