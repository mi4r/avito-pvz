package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mi4r/avito-pvz/internal/config"
	"github.com/mi4r/avito-pvz/internal/handler"
	auth "github.com/mi4r/avito-pvz/internal/middleware"
	"github.com/mi4r/avito-pvz/internal/storage"
)

func main() {
	cfg := config.NewConfig()

	// Инициализация подключения к БД
	pool := initDB(cfg)
	defer pool.Close()

	// Инициализация хранилищ
	userStorage := storage.NewUserStorage(pool)
	pvzStorage := storage.NewPVZStorage(pool)
	receptionStorage := storage.NewReceptionStorage(pool)
	productStorage := storage.NewProductStorage(pool)

	// Создание роутера
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Публичные маршруты
	r.Post("/dummyLogin", handler.DummyLogin(userStorage))
	r.Post("/register", handler.Register(userStorage))
	r.Post("/login", handler.Login(userStorage))

	// Защищенные маршруты
	r.Group(func(r chi.Router) {
		r.Use(auth.Auth)

		// PVZ endpoints
		r.Post("/pvz", handler.CreatePVZ(pvzStorage))
		r.Get("/pvz", handler.GetPVZs(pvzStorage))

		// Reception endpoints
		r.Post("/receptions", handler.CreateReception(receptionStorage))
		r.Post("/products", handler.AddProduct(productStorage, receptionStorage))
		r.Post("/pvz/{pvzId}/close_last_reception", handler.CloseLastReception(receptionStorage))
		r.Post("/pvz/{pvzId}/delete_last_product", handler.DeleteLastProduct(productStorage, receptionStorage))
	})

	// Запуск сервера
	port := getEnv("PORT", "8080")
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initDB(cfg config.Config) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// dbURL := cfg.GetDSN()
	dbURL := "postgres://mi4r:1234@localhost:5432/pvz_storage?sslmode=disable"
	// log.Println(dbURL)
	if dbURL == "" {
		dbURL = getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/pvz_storage?sslmode=disable")
	}
	// log.Println(dbURL)
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatal("Failed to parse database config:", err)
	}
	// log.Println(config.ConnString())
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("Failed to create connection pool:", err)
	}

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to database")
	return pool
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
