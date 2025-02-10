package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/adapters/db"
	updategrpc "yadro.com/course/update/adapters/grpc"
	"yadro.com/course/update/adapters/words"
	"yadro.com/course/update/adapters/xkcd"
	"yadro.com/course/update/config"
	"yadro.com/course/update/core"
)

func main() {

	// config
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	// logger
	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// database adapter
	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	if err := storage.Migrate(); err != nil {
		log.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	// xkcd adapter
	xkcd, err := xkcd.NewClient(cfg.XKCD.URL, cfg.XKCD.Timeout, log)
	if err != nil {
		log.Error("failed create XKCD client", "error", err)
		os.Exit(1)
	}

	// words adapter
	words, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("failed create Words client", "error", err)
		os.Exit(1)
	}

	// service
	updater, err := core.NewService(log, storage, xkcd, words, cfg.XKCD.Concurrency)
	if err != nil {
		log.Error("failed create Update service", "error", err)
		os.Exit(1)
	}

	// grpc server
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	updatepb.RegisterUpdateServer(s, updategrpc.NewServer(updater))
	reflection.Register(s)

	// context for Ctrl-C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		s.GracefulStop()
	}()

	if err := s.Serve(listener); err != nil {
		log.Error("failed to serve", "erorr", err)
		os.Exit(1)
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
