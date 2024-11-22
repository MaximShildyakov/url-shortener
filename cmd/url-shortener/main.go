package main

import (
	"errors"
	"net/http"
	// "log/slog"
	"os"
	"strings"

	mwLogger "github.com/MaximShildyakov/url-shortener/cmd/internal/http-server/middleware/logger"
	"github.com/MaximShildyakov/url-shortener/cmd/internal/http-server/middleware/logger/handlers/url/save"

	"github.com/MaximShildyakov/url-shortener/cmd/internal/config"
	"github.com/MaximShildyakov/url-shortener/cmd/internal/lib/logger/sl"
	storagePkg "github.com/MaximShildyakov/url-shortener/cmd/internal/storage"
	"github.com/MaximShildyakov/url-shortener/cmd/internal/storage/sqlite"

	// "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main(){

	// if os.Getenv("CONFIG_PATH") == "" {
    //     os.Setenv("CONFIG_PATH", "./config/local.yaml")  // или другой путь к вашему конфигурационному файлу
    // }
    
    cfg := config.MustLoad()

	// fmt.Println(cfg)

	log := setupLogger(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil{
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	id, err := storage.SaveURL("https://google.com", "google")
	if err != nil {
		if errors.Is(err, storagePkg.ErrURLExists) {
			log.Warn("URL already exists in the database", sl.Err(err))
		} else {
			log.Error("failed to save url", sl.Err(err))
			return
		}
	} else {
		log.Info("saved url", slog.Int64("id", id))
	}

	id, err = storage.SaveURL("https://new-url.com", "new")
	if err != nil {
		if strings.Contains(err.Error(), "url exists") {
			log.Warn("URL already exists in the database", sl.Err(err))
		} else {
			log.Error("failed to save url", sl.Err(err))
			return
		}
	} else {
		log.Info("saved url", slog.Int64("id", id))
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	// router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))
	
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timout,
		WriteTimeout: cfg.HTTPServer.Timout,
		IdleTimeout:  cfg.HTTPServer.IdleTimout,
	}

	if err := srv.ListenAndServe(); err != nil{
		log.Error("failed to start server")
	}

	log.Error("server stopped")

	// id, err = storage.SaveURL("https://google.com", "google")
	// if err != nil {
	// 	if strings.Contains(err.Error(), "url exists") {
	// 		log.Warn("URL already exists in the database", sl.Err(err))
	// 	} else {
	// 		log.Error("failed to save url", sl.Err(err))
	// 		return
	// 	}
	// } else {
	// 	log.Info("saved url", slog.Int64("id", id))
	// }

	// id, err := storage.SaveURL("https://google.com", "google")
	// if err != nil{
	// 	log.Error("failed to save url", sl.Err(err))
	// 	os.Exit(1)
	// }

	// log.Info("saved url", slog.Int64("id", id))

	// id, err = storage.SaveURL("https://google.com", "google")
	// if err != nil{
	// 	log.Error("failed to save url", sl.Err(err))
	// 	os.Exit(1)
	// }

	



	_ = storage

}

func setupLogger(env string) *slog.Logger{
	var log *slog.Logger

	switch env{
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	
	}

	return log
}