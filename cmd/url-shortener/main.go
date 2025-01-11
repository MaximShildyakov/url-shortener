package main

import (
	// "context"
	"net/http"
	"os"
	// "os/signal"
	// "syscall"
	// "time"
	// "strings"
	// "errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"



	"github.com/MaximShildyakov/url-shortener/internal/config"
	"github.com/MaximShildyakov/url-shortener/internal/http-server/handlers/redirect"
	"github.com/MaximShildyakov/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/MaximShildyakov/url-shortener/internal/http-server/middleware/logger"
	"github.com/MaximShildyakov/url-shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/MaximShildyakov/url-shortener/internal/lib/logger/sl"
	"github.com/MaximShildyakov/url-shortener/internal/storage/sqlite"
	//storagePkg "github.com/MaximShildyakov/url-shortener/internal/storage"
	"github.com/MaximShildyakov/url-shortener/internal/http-server/handlers/delete"
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

	log.Info("starting url-shortener",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil{
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	// id, err := storage.SaveURL("https://google.com", "google")
	// if err != nil {
	// 	if errors.Is(err, storagePkg.ErrURLExists) {
	// 		log.Warn("URL already exists in the database", sl.Err(err))
	// 	} else {
	// 		log.Error("failed to save url", sl.Err(err))
	// 		return
	// 	}
	// } else {
	// 	log.Info("saved url", slog.Int64("id", id))
	// }

	// id, err = storage.SaveURL("https://new-url.com", "new")
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

	

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	// router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router){
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
	})

	// router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))
	//router.Delete("/{alias}", delete.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	log.Info("HTTPServer config: %+v\n", cfg.HTTPServer)
	
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
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

	



	// _ = storage

}

func setupLogger(env string) *slog.Logger{
	var log *slog.Logger

	switch env{
	case envLocal:
		log = setupPrettySlog()
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

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}