package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/lucatorrekens/bakery-app/internal/api"
	appmw "github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-do-not-use-in-production"
	}

	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	r := chi.NewRouter()

	// --- Global middleware ---
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(appmw.InputSanitizer)

	// --- Health check (public) ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// --- Initialize repositories ---
	bakeryRepo := memory.NewBakeryRepo()
	orderRepo := memory.NewOrderRepo()
	reservationRepo := memory.NewReservationRepo()
	userRepo := memory.NewUserRepo()
	recurringOrderRepo := memory.NewRecurringOrderRepo()

	// --- Initialize services ---
	bakerySvc := service.NewBakeryService(bakeryRepo)

	authSvc := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:  userRepo,
		JWTSecret: jwtSecret,
	})

	paymentGateway := payment.NewStubGateway()
	paymentSvc := payment.NewPaymentService(payment.ServiceConfig{
		Gateway:   paymentGateway,
		OrderRepo: orderRepo,
	})

	orderSvc := service.NewOrderService(service.OrderServiceConfig{
		OrderRepo:  orderRepo,
		BakeryRepo: bakeryRepo,
		UserRepo:   userRepo,
		PaymentSvc: paymentSvc,
	})

	reservationSvc := service.NewReservationService(service.ReservationServiceConfig{
		BakeryRepo:      bakeryRepo,
		ReservationRepo: reservationRepo,
	})

	// --- Initialize handlers ---
	bakeryHandler := api.NewBakeryHandler(bakerySvc)
	orderHandler := api.NewOrderHandler(orderSvc)
	reservationHandler := api.NewReservationHandler(reservationSvc)
	paymentHandler := api.NewPaymentHandler(paymentSvc, orderRepo)
	authHandler := api.NewAuthHandler(authSvc)

	sellerSvc := service.NewSellerService(service.SellerServiceConfig{
		BakeryRepo:      bakeryRepo,
		OrderRepo:       orderRepo,
		ReservationRepo: reservationRepo,
	})
	sellerHandler := api.NewSellerHandler(sellerSvc)

	recurringOrderSvc := service.NewRecurringOrderService(service.RecurringOrderServiceConfig{
		RecurringRepo: recurringOrderRepo,
		BakeryRepo:    bakeryRepo,
	})
	recurringHandler := api.NewRecurringHandler(recurringOrderSvc)

	userSvc := service.NewUserService(userRepo)
	userHandler := api.NewUserHandler(userSvc)

	// --- Rate limiter for order/reservation submission ---
	rateLimiter := appmw.NewRateLimiter(appmw.RateLimitConfig{
		MaxRequests: 10,
		Window:      60_000_000_000, // 60 seconds in nanoseconds (time.Minute)
		UserIDExtractor: func(r *http.Request) string {
			return appmw.GetUserIDFromContext(r.Context())
		},
	})

	// --- Public routes (no JWT required) ---
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// Bakery browsing endpoints are public (no auth required)
	bakeryHandler.RegisterRoutes(r)

	// --- Protected API routes (require JWT auth) ---
	// Order, reservation, and payment endpoints require authentication.
	r.Group(func(r chi.Router) {
		r.Use(appmw.JWTAuth(jwtSecret))

		// Order endpoints — POST is rate-limited
		r.With(rateLimiter.Middleware).Post("/api/orders", orderHandler.CreateOrder)
		r.Get("/api/orders", orderHandler.ListOrders)
		r.Delete("/api/orders/{id}", orderHandler.DeleteOrder)

		// Reservation endpoints — POST is rate-limited
		r.With(rateLimiter.Middleware).Post("/api/reservations", reservationHandler.CreateReservation)
		r.Delete("/api/reservations/{id}", reservationHandler.DeleteReservation)

		// Payment callback
		paymentHandler.RegisterRoutes(r)

		// Seller portal routes (role check is done in handler)
		sellerHandler.RegisterRoutes(r)

		// Recurring order routes
		recurringHandler.RegisterRoutes(r)

		// User profile and holiday mode routes
		r.Get("/api/user/profile", userHandler.GetProfile)
		r.Put("/api/user/holiday", userHandler.UpdateHoliday)
		r.Get("/api/user/favorites", userHandler.GetFavorites)
		r.Put("/api/user/favorites", userHandler.UpdateFavorites)
	})

	log.Printf("Starting bakery-app server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
