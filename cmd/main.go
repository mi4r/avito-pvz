package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/mi4r/avito-pvz/internal/handler"
	auth "github.com/mi4r/avito-pvz/internal/middleware"
	"github.com/mi4r/avito-pvz/internal/storage"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// cfg := config.NewConfig()

	// Инициализация подключения к БД
	dbURL := "postgres://mi4r:1234@localhost:5432/pvz_storage?sslmode=disable"
	// dbURL := cfg.GetDSN()
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println(dbURL)
	store := storage.NewPostgresStorage(db)
	store.Migrate(dbURL)

	// Создание роутера
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Публичные маршруты
	r.Post("/dummyLogin", handler.DummyLogin())
	r.Post("/register", handler.Register(store))
	r.Post("/login", handler.Login(store))

	// Защищенные маршруты
	r.Group(func(r chi.Router) {
		r.Use(auth.Auth)

		// PVZ endpoints
		r.Post("/pvz", handler.CreatePVZ(store))
		r.Get("/pvz", handler.GetPVZs(store))

		// Reception endpoints
		r.Post("/receptions", handler.CreateReception(store))
		r.Post("/products", handler.AddProduct(store))
		r.Post("/pvz/{pvzId}/close_last_reception", handler.CloseLastReception(store))
		r.Post("/pvz/{pvzId}/delete_last_product", handler.DeleteLastProduct(store))
	})

	// Запуск сервера
	port := getEnv("PORT", "8080")
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
