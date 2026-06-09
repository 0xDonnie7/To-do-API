package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
	"todoListAPI/internal/data"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	jwt struct {
		secret string
	}
}

type application struct {
	cfg    config
	models data.Models
	logger *slog.Logger
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	var cfg config
	flag.IntVar(&cfg.port, "port", 8080, "server port number")
	flag.StringVar(&cfg.env, "env", "development", "environment(development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_URL"), "Postgres connection string")
	flag.IntVar(&cfg.db.MaxOpenConns, "db-max-open-conns", 50, "Postgres max open connections")
	flag.StringVar(&cfg.db.MaxIdleTime, "db-max-idle-time", "15m", "Postgres max idle time")
	flag.IntVar(&cfg.db.MaxIdleConns, "db-max-idle-conns", 25, "Postgres max idle connnections")
	flag.StringVar(&cfg.jwt.secret, "jwt-secret", os.Getenv("JWT-SECRET"), "JWT signing secret")

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		fmt.Println("failed to connect to database:", err)
		return
	}

	app := &application{
		cfg:    cfg,
		logger: logger,
		models: data.AllModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		app.logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.MaxOpenConns)
	db.SetMaxIdleConns(cfg.db.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.db.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
