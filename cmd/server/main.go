package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/lucatorrekens/bakery-app/internal/api"
	"github.com/lucatorrekens/bakery-app/internal/auth"
	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/email"
	"github.com/lucatorrekens/bakery-app/internal/invoice"
	appmw "github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/lucatorrekens/bakery-app/internal/notification"
	"github.com/lucatorrekens/bakery-app/internal/payment"
	"github.com/lucatorrekens/bakery-app/internal/push"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
	"github.com/lucatorrekens/bakery-app/internal/repository/postgres"
	"github.com/lucatorrekens/bakery-app/internal/service"
	"github.com/lucatorrekens/bakery-app/internal/upload"
	"github.com/lucatorrekens/bakery-app/internal/ws"
	"github.com/pressly/goose/v3"
)

func main() {
	// Load .env file if it exists (won't error if missing)
	_ = godotenv.Load()

	// Initialize Sentry error tracking (no-op if DSN is empty)
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      os.Getenv("APP_ENV"),
			Release:          os.Getenv("APP_VERSION"),
			TracesSampleRate: 0.1,
		})
		if err != nil {
			log.Printf("[SENTRY] init failed: %v", err)
		} else {
			log.Println("[SENTRY] initialized")
			defer sentry.Flush(2 * time.Second)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	appEnv := os.Getenv("APP_ENV")
	if jwtSecret == "" {
		if appEnv == "production" {
			log.Fatal("FATAL: JWT_SECRET must be set in production. Refusing to start with default secret.")
		}
		jwtSecret = "dev-secret-do-not-use-in-production"
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET env var for production!")
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
		contactEmail = "admin@maboulangerie.com"
	}

	r := chi.NewRouter()

	// --- Global middleware ---
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(appmw.SentryMiddleware())
	r.Use(chimw.RequestID)
	r.Use(appmw.SecurityHeaders)
	r.Use(appmw.CORS(appmw.CORSConfig{AllowedOrigin: frontendOrigin}))
	r.Use(appmw.BodyLimit(appmw.DefaultBodyLimit))

	// --- Image upload storage ---
	var uploadStorage upload.Storage
	switch os.Getenv("UPLOAD_STORAGE") {
	case "", "local":
		localStorage, err := upload.NewLocalStorage("./uploads", "/uploads")
		if err != nil {
			log.Fatalf("Failed to initialize upload storage: %v", err)
		}
		uploadStorage = localStorage
		log.Println("Upload storage: local (./uploads)")
	case "s3":
		log.Fatalf("FATAL: %v", upload.ErrS3NotImplemented)
	default:
		log.Fatalf("FATAL: unknown UPLOAD_STORAGE value %q. Supported: \"local\" (or leave unset).", os.Getenv("UPLOAD_STORAGE"))
	}

	// Serve uploaded files as static assets
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// --- Health check (public) ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// --- Initialize repositories ---
	databaseURL := os.Getenv("DATABASE_URL")

	var (
		bakeryRepo         domain.BakeryRepository
		orderRepo          domain.OrderRepository
		reservationRepo    domain.ReservationRepository
		userRepo           domain.UserRepository
		recurringOrderRepo domain.RecurringOrderRepository
		tokenRepo          domain.RegistrationTokenRepository
		bundleRepo         domain.BundleRepository
		reviewRepo         domain.ReviewRepository
		b2bRepo            domain.B2BRepository
		socialLoginRepo    domain.SocialLoginRepository
		payoutRepo         domain.PayoutRepository
	)

	if databaseURL != "" {
		pool, err := postgres.NewPool(context.Background(), postgres.Config{DatabaseURL: databaseURL})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer pool.Close()
		log.Println("Database: PostgreSQL (connected)")

		if err := runMigrations(pool); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Database migrations applied")

		bakeryRepo = postgres.NewBakeryRepo(pool)
		orderRepo = postgres.NewOrderRepo(pool)
		reservationRepo = postgres.NewReservationRepo(pool)
		userRepo = postgres.NewUserRepo(pool)
		recurringOrderRepo = postgres.NewRecurringOrderRepo(pool)
		tokenRepo = postgres.NewTokenRepo(pool)
		bundleRepo = postgres.NewBundleRepo(pool)
		reviewRepo = postgres.NewReviewRepo(pool)
		b2bRepo = postgres.NewB2BRepo(pool)
		socialLoginRepo = postgres.NewSocialLoginRepo(pool)
		payoutRepo = postgres.NewPayoutRepo(pool)
	} else {
		log.Println("Database: In-memory (set DATABASE_URL for PostgreSQL)")

		memBakeryRepo := memory.NewBakeryRepo()
		memOrderRepo := memory.NewOrderRepo()
		memReservationRepo := memory.NewReservationRepo()
		memUserRepo := memory.NewUserRepo()
		memRecurringOrderRepo := memory.NewRecurringOrderRepo()
		memTokenRepo := memory.NewTokenRepo()
		memSocialLoginRepo := memory.NewSocialLoginRepo()

		bakeryRepo = memBakeryRepo
		orderRepo = memOrderRepo
		reservationRepo = memReservationRepo
		userRepo = memUserRepo
		recurringOrderRepo = memRecurringOrderRepo
		tokenRepo = memTokenRepo
		socialLoginRepo = memSocialLoginRepo
		payoutRepo = memory.NewPayoutRepo()

		// Seed demo data only for in-memory mode
		seedDemoData(memBakeryRepo, memUserRepo, memOrderRepo, memReservationRepo, memRecurringOrderRepo, memTokenRepo)
	}

	// --- Initialize services ---
	bakerySvc := service.NewBakeryService(bakeryRepo)

	authSvc := service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     userRepo,
		TokenRepo:    tokenRepo,
		JWTSecret:    jwtSecret,
		ContactEmail: contactEmail,
	})

	// --- OAuth providers (only active when env vars are configured) ---
	oauthProviders := make(map[string]auth.OAuthProvider)
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientID != "" && googleClientSecret != "" {
		oauthProviders["google"] = &auth.GoogleProvider{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
		}
		log.Println("OAuth: Google provider enabled")
	}

	appleClientID := os.Getenv("APPLE_CLIENT_ID")
	appleTeamID := os.Getenv("APPLE_TEAM_ID")
	appleKeyID := os.Getenv("APPLE_KEY_ID")
	applePrivateKey := os.Getenv("APPLE_PRIVATE_KEY")
	if appleClientID != "" && appleTeamID != "" && appleKeyID != "" && applePrivateKey != "" {
		oauthProviders["apple"] = &auth.AppleProvider{
			ClientID:   appleClientID,
			TeamID:     appleTeamID,
			KeyID:      appleKeyID,
			PrivateKey: applePrivateKey,
		}
		log.Println("OAuth: Apple provider enabled")
	}

	oauthRedirectBase := os.Getenv("OAUTH_REDIRECT_BASE")
	if oauthRedirectBase == "" {
		oauthRedirectBase = frontendOrigin
	}

	var oauthSvc *service.OAuthService
	if len(oauthProviders) > 0 {
		oauthSvc = service.NewOAuthService(service.OAuthServiceConfig{
			Providers:       oauthProviders,
			SocialLoginRepo: socialLoginRepo,
			UserRepo:        userRepo,
			JWTSecret:       jwtSecret,
			RedirectBase:    oauthRedirectBase,
		})
	}

	var paymentGateway payment.PaymentGateway
	var stripeCustomerSvc *payment.StripeCustomerService
	if paymentMode == "stripe" {
		stripeKey := os.Getenv("STRIPE_SECRET_KEY")
		if stripeKey == "" {
			log.Fatal("STRIPE_SECRET_KEY is required when PAYMENT_GATEWAY=stripe")
		}
		paymentGateway = payment.NewStripeGateway(payment.StripeConfig{
			SecretKey:  stripeKey,
			SuccessURL: frontendOrigin + "/schedule?payment=success",
			CancelURL:  frontendOrigin + "/schedule?payment=cancelled",
			UserRepo:   userRepo,
			OrderRepo:  orderRepo,
		})
		stripeCustomerSvc = payment.NewStripeCustomerService(stripeKey, userRepo)
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

	// --- WebSocket hub for real-time notifications ---
	wsHub := ws.NewHub()

	// --- Push notification sender (Web Push via VAPID) ---
	vapidPublic := os.Getenv("VAPID_PUBLIC_KEY")
	vapidPrivate := os.Getenv("VAPID_PRIVATE_KEY")
	var pushSender *push.Sender
	var pushStore *push.Store
	if vapidPublic != "" && vapidPrivate != "" {
		pushStore = push.NewStore()
		pushSender = push.NewSender(vapidPublic, vapidPrivate, contactEmail, pushStore)
		log.Println("Push notifications: enabled")
	} else {
		log.Println("Push notifications: disabled (set VAPID_PUBLIC_KEY + VAPID_PRIVATE_KEY)")
	}

	// --- Invoice store & notification service ---
	invoiceStore := invoice.NewStore()
	notificationSvc := notification.NewService(notification.ServiceConfig{
		EmailSender:     emailSender,
		InvoiceStore:    invoiceStore,
		OrderRepo:       orderRepo,
		BakeryRepo:      bakeryRepo,
		UserRepo:        userRepo,
		ReservationRepo: reservationRepo,
		WSHub:           wsHub,
		PushSender:      pushSender,
	})

	paymentSvc := payment.NewPaymentService(payment.ServiceConfig{
		Gateway:          paymentGateway,
		OrderRepo:        orderRepo,
		OnOrderConfirmed: notificationSvc.OnPaymentConfirmed,
	})

	// --- Stripe Connect payout service (only when Stripe is active) ---
	// Created before orderSvc so OnOrderRefunded can be wired into the order cancellation flow.
	var payoutHandler *api.PayoutHandler
	var payoutSvc *service.PayoutService
	if paymentMode == "stripe" {
		stripeKey := os.Getenv("STRIPE_SECRET_KEY")
		platformID := os.Getenv("STRIPE_CONNECT_PLATFORM_ID")
		connectSvc := payment.NewConnectService(payment.ConnectConfig{
			StripeKey:      stripeKey,
			PlatformAcctID: platformID,
		})
		payoutSvc = service.NewPayoutService(service.PayoutServiceConfig{
			ConnectSvc: connectSvc,
			PayoutRepo: payoutRepo,
			BakeryRepo: bakeryRepo,
			OrderRepo:  orderRepo,
		})
		payoutHandler = api.NewPayoutHandler(payoutSvc, bakeryRepo)
	}

	// Wire OnOrderRefunded: when the order service refunds a captured payment, reverse the payout
	var onOrderRefunded func(ctx context.Context, orderID string) error
	if payoutSvc != nil {
		onOrderRefunded = payoutSvc.OnOrderRefunded
	}

	orderSvc := service.NewOrderService(service.OrderServiceConfig{
		OrderRepo:        orderRepo,
		BakeryRepo:       bakeryRepo,
		UserRepo:         userRepo,
		PaymentSvc:       paymentSvc,
		PaymentGateway:   paymentGateway,
		OnOrderCancelled: notificationSvc.OnOrderCancelled,
		OnOrderRefunded:  onOrderRefunded,
		OnNewOrder:       notificationSvc.OnNewOrder,
	})

	reservationSvc := service.NewReservationService(service.ReservationServiceConfig{
		BakeryRepo:      bakeryRepo,
		ReservationRepo: reservationRepo,
		Notifications:   notificationSvc,
	})

	// --- Initialize handlers ---
	bakeryHandler := api.NewBakeryHandler(bakerySvc)
	orderHandler := api.NewOrderHandler(orderSvc)
	reservationHandler := api.NewReservationHandler(reservationSvc)
	paymentHandler := api.NewPaymentHandler(paymentSvc, orderRepo, paymentMode)
	authHandler := api.NewAuthHandler(authSvc)
	invoiceHandler := api.NewInvoiceHandler(invoiceStore, orderRepo)

	var oauthHandler *api.OAuthHandler
	if oauthSvc != nil {
		oauthHandler = api.NewOAuthHandler(oauthSvc, []byte(jwtSecret))
	}

	uploadHandler := api.NewUploadHandler(uploadStorage)

	sellerSvc := service.NewSellerService(service.SellerServiceConfig{
		BakeryRepo:      bakeryRepo,
		OrderRepo:       orderRepo,
		ReservationRepo: reservationRepo,
		PaymentGateway:  paymentGateway,
		Notifications:   notificationSvc,
	})
	sellerHandler := api.NewSellerHandler(sellerSvc)

	// Wire payout into order delivery lifecycle via the seller service
	if payoutSvc != nil {
		sellerSvc.OnOrderDelivered = payoutSvc.OnOrderDelivered
	}

	recurringOrderSvc := service.NewRecurringOrderService(service.RecurringOrderServiceConfig{
		RecurringRepo: recurringOrderRepo,
		BakeryRepo:    bakeryRepo,
	})
	recurringHandler := api.NewRecurringHandler(recurringOrderSvc)

	// --- B2B service and handler (requires PostgreSQL) ---
	var b2bHandler *api.B2BHandler
	if b2bRepo != nil {
		b2bSvc := service.NewB2BService(service.B2BServiceConfig{
			B2BRepo:    b2bRepo,
			UserRepo:   userRepo,
			BakeryRepo: bakeryRepo,
			OrderRepo:  orderRepo,
			JWTSecret:  jwtSecret,
		})
		b2bHandler = api.NewB2BHandler(b2bSvc, bakeryRepo)
	}

	// --- Review service and handler ---
	var reviewHandler *api.ReviewHandler
	if reviewRepo != nil {
		reviewSvc := service.NewReviewService(service.ReviewServiceConfig{
			ReviewRepo: reviewRepo,
			OrderRepo:  orderRepo,
			BakeryRepo: bakeryRepo,
		})
		reviewHandler = api.NewReviewHandler(reviewSvc, userRepo)
	}

	userSvc := service.NewUserServiceFull(service.UserServiceConfig{
		UserRepo:           userRepo,
		OrderRepo:          orderRepo,
		ReservationRepo:    reservationRepo,
		RecurringOrderRepo: recurringOrderRepo,
		ReviewRepo:         reviewRepo,
		SocialLoginRepo:    socialLoginRepo,
		B2BRepo:            b2bRepo,
		PushStore:          pushStore,
		StripeCustomerSvc:  stripeCustomerSvc,
	})
	userHandler := api.NewUserHandler(userSvc)

	// --- Payment method handler (only when Stripe is active) ---
	var paymentMethodHandler *api.PaymentMethodHandler
	if stripeCustomerSvc != nil {
		paymentMethodHandler = api.NewPaymentMethodHandler(stripeCustomerSvc)
	}

	// --- Push notification handler (only when VAPID keys are configured) ---
	var pushHandler *api.PushHandler
	if pushSender != nil {
		pushHandler = api.NewPushHandler(pushSender, pushStore)
	}

	// --- Bundle service and handler (requires bundleRepo) ---
	var bundleSvc domain.BundleService
	var bundleHandler *api.BundleHandler
	if bundleRepo != nil {
		bundleSvc = service.NewBundleService(service.BundleServiceConfig{
			Repo:       bundleRepo,
			BakeryRepo: bakeryRepo,
			Hub:        wsHub,
		})
		bundleHandler = api.NewBundleHandler(bundleSvc, bakeryRepo, bundleRepo, wsHub)
	}

	// --- Rate limiter for order/reservation submission ---
	rateLimiter := appmw.NewRateLimiter(appmw.RateLimitConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		UserIDExtractor: func(r *http.Request) string {
			return appmw.GetUserIDFromContext(r.Context())
		},
	})

	// --- IP-based rate limiter for auth endpoints (brute-force protection) ---
	authRateLimiter := appmw.NewRateLimiter(appmw.RateLimitConfig{
		MaxRequests: 5,
		Window:      time.Minute,
		UserIDExtractor: func(r *http.Request) string {
			// Use RemoteAddr as the primary key (set by trusted reverse proxy or
			// Go's net/http from the TCP connection). Only fall back to
			// X-Forwarded-For when a trusted proxy list is configured.
			return r.RemoteAddr
		},
	})

	// --- Public routes (no JWT required) ---
	r.With(authRateLimiter.Middleware).Post("/api/auth/register", authHandler.Register)
	r.With(authRateLimiter.Middleware).Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/request-access", authHandler.RequestAccess)

	// OAuth social login routes (public — initiate and callback)
	if oauthHandler != nil {
		r.Get("/api/auth/oauth/{provider}", oauthHandler.GetAuthURL)
		r.Post("/api/auth/oauth/{provider}/callback", oauthHandler.HandleCallback)
	}

	// WebSocket endpoint — auth via token query param during upgrade
	r.Get("/api/ws", wsHub.HandleUpgrade(jwtSecret))

	// VAPID public key endpoint — public (browser needs it to subscribe)
	if pushHandler != nil {
		r.Get("/api/push/vapid-key", pushHandler.GetVAPIDKey)
	}

	// Bakery browsing endpoints are public (no auth required)
	bakeryHandler.RegisterRoutes(r)

	// B2B Comptoir portal routes (handles its own auth internally)
	if b2bHandler != nil {
		b2bHandler.RegisterRoutes(r, jwtSecret, userRepo)
	}

	// Review listing is public (no auth required)
	if reviewHandler != nil {
		reviewHandler.RegisterPublicRoutes(r)
	}

	// Bundle browsing endpoints (list, get, impact) are public
	if bundleHandler != nil {
		r.Get("/api/bundles/impact", bundleHandler.GetImpact)
		r.Get("/api/bundles", bundleHandler.ListBundles)
		r.Get("/api/bundles/{id}", bundleHandler.GetBundle)
	}

	// --- Protected API routes (require JWT auth) ---
	// Order, reservation, and payment endpoints require authentication.
	r.Group(func(r chi.Router) {
		r.Use(appmw.JWTAuth(jwtSecret, userRepo))

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

		// Payout and Stripe Connect routes (role check is done in handler)
		if payoutHandler != nil {
			payoutHandler.RegisterRoutes(r)
		}

		// Review routes (create, report — auth required; hide — seller role checked in handler)
		if reviewHandler != nil {
			reviewHandler.RegisterRoutes(r)
			reviewHandler.RegisterSellerRoutes(r)
		}

		// Image upload (role check is done in handler — seller or admin)
		uploadHandler.RegisterRoutes(r)

		// Recurring order routes
		recurringHandler.RegisterRoutes(r)

		// Invoice retrieval (owner-only access enforced in handler)
		invoiceHandler.RegisterRoutes(r)

		// User profile and holiday mode routes
		r.Get("/api/user/profile", userHandler.GetProfile)
		r.Put("/api/user/holiday", userHandler.UpdateHoliday)
		r.Get("/api/user/favorites", userHandler.GetFavorites)
		r.Put("/api/user/favorites", userHandler.UpdateFavorites)
		r.Get("/api/user/data-export", userHandler.DataExport)
		r.Delete("/api/user/account", userHandler.DeleteAccount)

		// Saved payment methods (only when Stripe is active)
		if paymentMethodHandler != nil {
			paymentMethodHandler.RegisterRoutes(r)
		}

		// Push notification subscription management (only when VAPID keys are configured)
		if pushHandler != nil {
			r.Post("/api/user/push/subscribe", pushHandler.Subscribe)
			r.Delete("/api/user/push/unsubscribe", pushHandler.Unsubscribe)
		}

		// Bundle mutation routes (reserve, confirm, cancel, create, publish)
		if bundleHandler != nil {
			r.Post("/api/bundles", bundleHandler.CreateBundle)
			r.Post("/api/bundles/{id}/publish", bundleHandler.PublishBundle)
			r.Post("/api/bundles/{id}/reserve", bundleHandler.ReserveBundle)
			r.Post("/api/bundles/{id}/reserve/confirm", bundleHandler.ConfirmReservation)
			r.Delete("/api/bundle-reservations/{id}", bundleHandler.CancelReservation)
		}
	})

	// --- Stripe webhook (public — Stripe needs to reach it without JWT) ---
	if paymentMode == "stripe" {
		webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
		if webhookSecret == "" {
			if appEnv == "production" {
				log.Fatal("FATAL: STRIPE_WEBHOOK_SECRET must be set in production when PAYMENT_GATEWAY=stripe.")
			}
			log.Println("WARNING: STRIPE_WEBHOOK_SECRET is empty. Webhook signature verification will fail.")
		}
		stripeWebhook := payment.NewStripeWebhookHandler(webhookSecret, paymentSvc, orderRepo)
		if payoutSvc != nil {
			stripeWebhook.SetPayoutReverser(payoutSvc)
		}
		r.Post("/api/stripe/webhook", stripeWebhook.HandleWebhook)
	}

	// --- Stripe Connect webhook (public — Stripe sends it without JWT) ---
	if connectWebhookSecret := os.Getenv("STRIPE_CONNECT_WEBHOOK_SECRET"); connectWebhookSecret != "" {
		connectWebhook := payment.NewConnectWebhookHandler(connectWebhookSecret, bakeryRepo)
		r.Post("/api/stripe/connect-webhook", connectWebhook.HandleWebhook)
	}

	// --- Context for background workers (cancelled on SIGINT/SIGTERM) ---
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- Start bundle expiration worker ---
	if bundleSvc != nil {
		go service.StartExpirationWorker(ctx, bundleSvc, wsHub)
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Printf("Starting bakery-app server on :%s", port)

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Forced shutdown: %v", err)
	}
	log.Println("Server stopped")
}

// runMigrations applies pending database migrations using goose.
func runMigrations(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	goose.SetDialect("postgres")
	return goose.Up(db, "db/migrations")
}
