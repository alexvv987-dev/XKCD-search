package main

import (
	"context"
	"errors"
	"flag"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"

	"yadro.com/course/web/config"
	"yadro.com/course/web/handlers"
)

const templatesGlob = "templates/*.html"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	apiClient := &http.Client{Timeout: cfg.HTTPConfig.Timeout}
	tmpl := template.Must(template.ParseGlob(templatesGlob))
	mux := http.NewServeMux()

	mux.Handle("GET /", handlers.NewSearchHandler(tmpl, log, cfg.APIAddress, apiClient))
	mux.Handle("GET /login", handlers.NewLoginPageHandler(tmpl, log, cfg.APIAddress, apiClient))
	mux.Handle("POST /login", handlers.NewLoginPageHandler(tmpl, log, cfg.APIAddress, apiClient))
	mux.Handle("GET /admin", handlers.NewAdminHandler(tmpl, log, cfg.APIAddress, apiClient))
	mux.Handle("GET /admin/api/status", handlers.NewAdminStatusAPIHandler(log, cfg.APIAddress, apiClient))
	mux.Handle("GET /admin/api/stats", handlers.NewAdminStatsAPIHandler(log, cfg.APIAddress, apiClient))
	mux.Handle("POST /admin/update", handlers.NewUpdateHandler(log, cfg.APIAddress, apiClient))
	mux.Handle("POST /admin/drop", handlers.NewDropHandler(log, cfg.APIAddress, apiClient))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})
	return slog.New(handler)
}
