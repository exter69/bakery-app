package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/lucatorrekens/bakery-app/internal/api"
	"github.com/lucatorrekens/bakery-app/internal/email"
	"github.com/lucatorrekens/bakery-app/internal/invoice"
	appmw "github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/notification"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/service"
)

func main() {
	// Load .env file if it exists (won't error if missing)
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-do-not-use-in-production"
	}
	if jwtSecret == "dev-secret-do-not-use-in-production" {
		log.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET env var for production!")
	}

	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	paymentMode := os.Getenv("PAYMENT_GATEWAY")
	if paymentMode == "" {
		paymentMode = "stub"
	}

	contactEmail := os.Getenv("CONTACT_EMAIL")
	if contactEmail == "" {
		contactEmail = "admin@mieetbeurre.com"
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
	tokenRepo := memory.NewTokenRepo()

	// --- Seed demo data ---
	seedDemoData(bakeryRepo, userRepo, orderRepo, reservationRepo, recurringOrderRepo, tokenRepo)

	// --- Initialize services ---
	bakerySvc := service.NewBakeryService(bakeryRepo)

	authSvc := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     userRepo,
		TokenRepo:    tokenRepo,
		JWTSecret:    jwtSecret,
		ContactEmail: contactEmail,
	})

	var paymentGateway payment.PaymentGateway
	if paymentMode == "stripe" {
		stripeKey := os.Getenv("STRIPE_SECRET_KEY")
		if stripeKey == "" {
			log.Fatal("STRIPE_SECRET_KEY is required when PAYMENT_GATEWAY=stripe")
		}
		paymentGateway = payment.NewStripeGateway(payment.StripeConfig{
			SecretKey:  stripeKey,
			SuccessURL: frontendOrigin + "/schedule?payment=success",
			CancelURL:  frontendOrigin + "/schedule?payment=cancelled",
		})
		log.Println("Payment gateway: Stripe")
	} else {
		paymentGateway = payment.NewStubGateway()
		log.Println("Payment gateway: Stub (dev mode)")
	}

	// --- Email sender (LogSender for dev, SMTPSender when SMTP env vars are present) ---
	var emailSender email.Sender
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost != "" {
		smtpPort := 587
		if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				smtpPort = p
			}
		}
		emailSender = &email.SMTPSender{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		}
		log.Println("Email sender: SMTP")
	} else {
		emailSender = &email.LogSender{}
		log.Println("Email sender: Log (dev mode — set SMTP_HOST to enable real emails)")
	}

	// --- Invoice store & notification service ---
	invoiceStore := invoice.NewStore()
	notificationSvc := notification.NewService(notification.ServiceConfig{
		EmailSender:  emailSender,
		InvoiceStore: invoiceStore,
		OrderRepo:    orderRepo,
		BakeryRepo:   bakeryRepo,
		UserRepo:     userRepo,
	})

	paymentSvc := payment.NewPaymentService(payment.ServiceConfig{
		Gateway:          paymentGateway,
		OrderRepo:        orderRepo,
		OnOrderConfirmed: notificationSvc.OnPaymentConfirmed,
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
	invoiceHandler := api.NewInvoiceHandler(invoiceStore, orderRepo)

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

	// --- IP-based rate limiter for auth endpoints (brute-force protection) ---
	authRateLimiter := appmw.NewRateLimiter(appmw.RateLimitConfig{
		MaxRequests: 5,
		Window:      60_000_000_000, // 60 seconds in nanoseconds (time.Minute)
		UserIDExtractor: func(r *http.Request) string {
			// Use IP address for unauthenticated endpoints
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.RemoteAddr
			}
			return ip
		},
	})

	// --- Public routes (no JWT required) ---
	r.Post("/api/auth/register", authHandler.Register)
	r.With(authRateLimiter.Middleware).Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/request-access", authHandler.RequestAccess)

	// Bakery browsing endpoints are public (no auth required)
	bakeryHandler.RegisterRoutes(r)

	// --- Protected API routes (require JWT auth) ---
	// Order, reservation, and payment endpoints require authentication.
	r.Group(func(r chi.Router) {
		r.Use(appmw.JWTAuth(jwtSecret))

		// Admin endpoints (role check done in handler)
		r.Post("/api/admin/tokens", authHandler.CreateToken)

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

		// Invoice retrieval (owner-only access enforced in handler)
		invoiceHandler.RegisterRoutes(r)

		// User profile and holiday mode routes
		r.Get("/api/user/profile", userHandler.GetProfile)
		r.Put("/api/user/holiday", userHandler.UpdateHoliday)
		r.Get("/api/user/favorites", userHandler.GetFavorites)
		r.Put("/api/user/favorites", userHandler.UpdateFavorites)
	})

	// --- Stripe webhook (public — Stripe needs to reach it without JWT) ---
	if paymentMode == "stripe" {
		webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
		stripeWebhook := payment.NewStripeWebhookHandler(webhookSecret, paymentSvc)
		r.Post("/api/stripe/webhook", stripeWebhook.HandleWebhook)
	}

	log.Printf("Starting bakery-app server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
