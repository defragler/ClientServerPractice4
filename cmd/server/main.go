package main

import (
	"log"
	"net/http"
	"time"

	"device-registry-api/internal/handler"
	appmw "device-registry-api/internal/middleware"
	"device-registry-api/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("=== Мережевий реєстр пристроїв (Device Registry API) ===")
	repo := repository.NewMemoryDeviceRepository()
	log.Println("[SERVER] Ініціалізація сховища даних (sync.RWMutex: active)")

	h := handler.NewDeviceHandler(repo)
	r := chi.NewRouter()
	r.Use(appmw.CORS)
	r.Use(func(next http.Handler) http.Handler {
		return appmw.Recoverer(next, func(w http.ResponseWriter, message string, status int) {
			handler.WriteProblemForMiddleware(w, status, message)
		})
	})
	r.Use(appmw.Logger)
	r.Route("/api/v1/devices", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})
	log.Println("[SERVER] Конвеєр Middleware підключено: Logger, Recoverer, CORS")
	log.Println("[SERVER] Маршрути Chi змонтовано на префіксі /api/v1/devices")

	srv := &http.Server{Addr: ":8080", Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Println("[SERVER] HTTP-сервер запущено на порту :8080...")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
