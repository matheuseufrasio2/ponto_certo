package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/matheuseufrasio2/ponto_certo/docs"
	"github.com/matheuseufrasio2/ponto_certo/internal/config"
	"github.com/matheuseufrasio2/ponto_certo/internal/database"
	"github.com/matheuseufrasio2/ponto_certo/internal/domain"
	"github.com/matheuseufrasio2/ponto_certo/internal/handler"
	internalMiddleware "github.com/matheuseufrasio2/ponto_certo/internal/middleware"
	"github.com/matheuseufrasio2/ponto_certo/internal/repository"
	"github.com/matheuseufrasio2/ponto_certo/internal/service"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Ponto Certo API
// @version 1.0
// @description High-performance RESTful API for barbershop scheduling and operations management.
// @BasePath /api/v1

// @contact.name Matheus
// @contact.email matheuseufrasio2@gmail.com

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Digite "Bearer " seguido pelo token JWT.
func main() {
	// 1. Load application configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup structured logging with slog
	setupLogger(cfg.Env)

	// 3. Initialize PostgreSQL connection pool
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()

	dbPool, err := database.NewPostgresPool(dbCtx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("❌ Database connection failure", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// 4. Instantiate repositories
	userRepo := repository.NewPostgresUserRepository(dbPool)
	catalogRepo := repository.NewPostgresCatalogRepository(dbPool)
	appointmentRepo := repository.NewPostgresAppointmentRepository(dbPool)
	billingRepo := repository.NewPostgresBillingRepository(dbPool)

	// 5. Instantiate services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	catalogService := service.NewCatalogService(catalogRepo)
	appointmentService := service.NewAppointmentService(appointmentRepo, catalogRepo)
	billingService := service.NewBillingService(billingRepo)

	// 6. Instantiate handlers
	authHandler := handler.NewAuthHandler(authService)
	catalogHandler := handler.NewCatalogHandler(catalogService)
	appointmentHandler := handler.NewAppointmentHandler(appointmentService)
	billingHandler := handler.NewBillingHandler(billingService)

	// 7. Initialize Chi router and standard middlewares
	r := chi.NewRouter()

	r.Use(middleware.RequestID) // Request tracing ID
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // Protect against panics

	// 8. Register API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Get("/health", handler.HealthCheck)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes requiring valid JWT
		r.Group(func(r chi.Router) {
			r.Use(internalMiddleware.AuthMiddleware(cfg.JWTSecret))

			// SUPER_ADMIN routes
			r.With(internalMiddleware.RequireRoles(domain.RoleSuperAdmin)).Get("/admin/stats", func(w http.ResponseWriter, r *http.Request) {
				_ = handler.WriteJSON(w, http.StatusOK, map[string]string{"message": "Acesso concedido ao Super Admin!"})
			})

			// Owner and Manager routes
			r.With(internalMiddleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleOrgOwner, domain.RoleUnitManager)).Get("/manager/overview", func(w http.ResponseWriter, r *http.Request) {
				_ = handler.WriteJSON(w, http.StatusOK, map[string]string{"message": "Acesso concedido à gestão!"})
			})

			// Unit catalog lookup (accessible by staff)
			r.With(internalMiddleware.RequireRoles(
				domain.RoleSuperAdmin,
				domain.RoleOrgOwner,
				domain.RoleUnitManager,
				domain.RoleReceptionist,
				domain.RoleBarber,
			)).Get("/units/{unit_id}/catalog", catalogHandler.GetUnitCatalog)

			// Reception & Unit Operations routes
			r.With(internalMiddleware.RequireRoles(
				domain.RoleSuperAdmin,
				domain.RoleOrgOwner,
				domain.RoleUnitManager,
				domain.RoleReceptionist,
			)).Route("/reception", func(r chi.Router) {
				r.Get("/appointments", appointmentHandler.ListByUnit)
				r.Post("/appointments", appointmentHandler.Create)
				r.Post("/checkout", billingHandler.Checkout)
			})

			// Barber operational routes
			r.With(internalMiddleware.RequireRoles(
				domain.RoleSuperAdmin,
				domain.RoleBarber,
			)).Route("/barber", func(r chi.Router) {
				r.Get("/my-queue", appointmentHandler.ListMyQueue)
				r.Post("/appointments/{id}/complete", appointmentHandler.CompleteCut)
			})
		})
	})

	// Swagger documentation route
	swaggerURL := fmt.Sprintf("http://localhost:%s/swagger/doc.json", cfg.Port)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(swaggerURL),
	))

	// 9. Configure HTTP server with production timeouts
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// 10. Graceful Shutdown: listen for OS termination signals (SIGINT, SIGTERM)
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("🚀 Server started", "port", cfg.Port, "env", cfg.Env)
		slog.Info("📚 Swagger documentation", "url", fmt.Sprintf("http://localhost:%s/swagger/index.html", cfg.Port))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Block until server error or shutdown signal received
	select {
	case err := <-serverErrors:
		slog.Error("❌ Server fatal error", "error", err)
		os.Exit(1)

	case <-shutdownCtx.Done():
		slog.Info("🛑 Shutting down server gracefully...")

		// Allow active requests up to 10 seconds to finish
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctxTimeout); err != nil {
			slog.Error("⚠️ Graceful shutdown failed", "error", err)
			_ = server.Close()
		}

		slog.Info("✅ Server stopped cleanly")
	}
}

// setupLogger initializes slog with JSON in production and readable text in development.
func setupLogger(env string) {
	var logHandler slog.Handler

	if env == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	slog.SetDefault(slog.New(logHandler))
}
