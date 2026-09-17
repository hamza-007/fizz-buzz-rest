package server

import (
	"net/http"

	fizzbuzz "fizz-buzz-rest/internal/fizzbuzz"
	stats "fizz-buzz-rest/internal/stats"
	rest "fizz-buzz-rest/server/rest"
	config "fizz-buzz-rest/utils/config"
	resp "fizz-buzz-rest/utils/resp"

	chi "github.com/go-chi/chi/v5"
	middleware "github.com/go-chi/chi/v5/middleware"
	cors "github.com/go-chi/cors"
	httprate "github.com/go-chi/httprate"
)

const maxRequestBytes = 1 << 16

const compressionLevel = 5

// TODO(lifetime): /api/v1 is a promise. A payload change ships as /api/v2 next
func routes() *chi.Mux {

	handlers := rest.New(rest.Options{
		MaxLimit: config.FizzBuzz().MaxLimit,
		Stats:    stats.NewCounter[fizzbuzz.Request](config.FizzBuzz().MaxStatsKeys),
		Health:   Health,
	})

	router := chi.NewRouter()

	router.Use(middleware.CleanPath)

	router.Use(middleware.RealIP)

	router.Use(requestID)
	router.Use(echoRequestID)

	router.Use(accessLog)
	router.Use(recoverPanic)

	crossOrigin := len(config.APP().CORS) > 0
	router.Use(secureHeaders(crossOrigin))
	if crossOrigin {
		router.Use(corsHandler())
	}

	router.Use(middleware.RequestSize(maxRequestBytes))
	router.Use(middleware.Compress(compressionLevel))
	if timeout := config.APP().RequestTimeout; timeout > 0 {
		router.Use(middleware.Timeout(timeout))
	}

	router.Route("/api/v1", func(api chi.Router) {

		if config.APP().RateLimit > 0 {
			api.Use(throttle())
		}

		api.Get("/fizzbuzz", handlers.FizzBuzz)
		api.Get("/stats", handlers.Stats)
	})

	router.Get("/health", handlers.Health)

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		resp.SendStatus(r.Context(), w, http.StatusNotFound, "not_found", "unknown endpoint")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allowedMethods())
		resp.SendStatus(r.Context(), w, http.StatusMethodNotAllowed, "method_not_allowed",
			"method "+r.Method+" is not supported on this endpoint")
	})

	return router
}

func allowedMethods() string {
	if len(config.APP().CORS) > 0 {
		return "GET, HEAD, OPTIONS"
	}
	return "GET, HEAD"
}

func corsHandler() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   config.APP().CORS,
		AllowedMethods:   []string{http.MethodGet, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Content-Type", middleware.RequestIDHeader},
		ExposedHeaders:   []string{middleware.RequestIDHeader},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

// TODO: counters are in-memory, so N replicas allow N times the rate; httprate-redis fixes it.
func throttle() func(http.Handler) http.Handler {
	return httprate.Limit(
		config.APP().RateLimit,
		config.APP().RateWindow,
		httprate.WithKeyByIP(),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			resp.SendStatus(r.Context(), w, http.StatusTooManyRequests, "rate_limited", "too many requests, retry later")
		}),
	)
}
