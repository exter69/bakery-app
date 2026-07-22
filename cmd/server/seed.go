package main

import (
	"context"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/repository/memory"
)

func hashPassword(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h)
}

func seedDemoData(
	bakeryRepo *memory.BakeryRepo,
	userRepo *memory.UserRepo,
	orderRepo *memory.OrderRepo,
	reservationRepo *memory.ReservationRepo,
	recurringRepo *memory.RecurringOrderRepo,
	tokenRepo *memory.TokenRepo,
) {
	ctx := context.Background()
	now := time.Now()

	log.Println("=== Seeding demo data ===")

	// ===================== USERS =====================
	// Admin (role 0)
	admin := &domain.User{
		ID: "user-admin", Username: "admin", PasswordHash: hashPassword("admin123"),
		Role: domain.RoleAdmin, CreatedAt: now,
	}
	_ = userRepo.Save(ctx, admin)

	// Sellers (role 1)
	seller1 := &domain.User{
		ID: "seller-1", Username: "baker_jean", PasswordHash: hashPassword("baker123"),
		Role: domain.RoleSeller, ContactEmail: "jean@laboulangerie.fr", CreatedAt: now,
	}
	_ = userRepo.Save(ctx, seller1)

	seller2 := &domain.User{
		ID: "seller-2", Username: "baker_marie", PasswordHash: hashPassword("baker123"),
		Role: domain.RoleSeller, ContactEmail: "marie@mieetbeurre.fr", CreatedAt: now,
	}
	_ = userRepo.Save(ctx, seller2)

	// Customers (role 2)
	customer1 := &domain.User{
		ID: "customer-1", Username: "alice", PasswordHash: hashPassword("customer123"),
		Role: domain.RoleCustomer, FavoriteProducts: []string{"prod-1-1", "prod-1-3", "prod-2-2"}, CreatedAt: now,
	}
	_ = userRepo.Save(ctx, customer1)

	customer2 := &domain.User{
		ID: "customer-2", Username: "bob", PasswordHash: hashPassword("customer123"),
		Role: domain.RoleCustomer, FavoriteProducts: []string{"prod-2-1", "prod-3-2"}, CreatedAt: now,
	}
	_ = userRepo.Save(ctx, customer2)

	log.Println("  ✓ Users: admin, baker_jean, baker_marie, alice, bob")

	// ===================== BAKERIES =====================
	bakeryRepo.SeedBakery(domain.Bakery{
		ID: "bakery-1", OwnerID: "seller-1", Name: "La Boulangerie du Coin",
		PhotoURL: "https://images.unsplash.com/photo-1509440159596-0249088772ff?w=600",
		Description: "Traditional French bakery with fresh bread baked every morning since 1987.",
		Address: "12 Rue de la Paix, Paris", Latitude: 48.8698, Longitude: 2.3298,
		Schedule: []domain.DaySchedule{
			{Day: domain.Monday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 19, Minute: 0}, IsOpen: true},
			{Day: domain.Tuesday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 19, Minute: 0}, IsOpen: true},
			{Day: domain.Wednesday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 19, Minute: 0}, IsOpen: true},
			{Day: domain.Thursday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 19, Minute: 0}, IsOpen: true},
			{Day: domain.Friday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 19, Minute: 0}, IsOpen: true},
			{Day: domain.Saturday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 30}, CloseTime: domain.TimeOfDay{Hour: 14, Minute: 0}, IsOpen: true},
			{Day: domain.Sunday, IsOpen: false},
		},
		MinDelivery: 1000, CreatedAt: now,
	})

	bakeryRepo.SeedBakery(domain.Bakery{
		ID: "bakery-2", OwnerID: "seller-2", Name: "Mie & Beurre",
		PhotoURL: "https://images.unsplash.com/photo-1517433670267-08bbd4be890f?w=600",
		Description: "Artisan sourdough and butter croissants made with organic flour.",
		Address: "45 Avenue des Champs, Lyon", Latitude: 45.7640, Longitude: 4.8357,
		Schedule: []domain.DaySchedule{
			{Day: domain.Monday, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 30}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Tuesday, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 30}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Wednesday, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 30}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Thursday, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 30}, CloseTime: domain.TimeOfDay{Hour: 18, Minute: 0}, IsOpen: true},
			{Day: domain.Friday, OpenTime: domain.TimeOfDay{Hour: 6, Minute: 30}, CloseTime: domain.TimeOfDay{Hour: 20, Minute: 0}, IsOpen: true},
			{Day: domain.Saturday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 16, Minute: 0}, IsOpen: true},
			{Day: domain.Sunday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 13, Minute: 0}, IsOpen: true},
		},
		MinDelivery: 1000, CreatedAt: now,
	})

	bakeryRepo.SeedBakery(domain.Bakery{
		ID: "bakery-3", OwnerID: "seller-1", Name: "Le Fournil de Max",
		PhotoURL: "https://images.unsplash.com/photo-1556471013-0001958d2f12?w=600",
		Description: "Small family bakery specializing in wood-fired bread and seasonal pastries.",
		Address: "8 Place du Marché, Bordeaux", Latitude: 44.8378, Longitude: -0.5792,
		Schedule: []domain.DaySchedule{
			{Day: domain.Monday, IsOpen: false},
			{Day: domain.Tuesday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 13, Minute: 30}, IsOpen: true},
			{Day: domain.Wednesday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 13, Minute: 30}, IsOpen: true},
			{Day: domain.Thursday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 13, Minute: 30}, IsOpen: true},
			{Day: domain.Friday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 13, Minute: 30}, IsOpen: true},
			{Day: domain.Saturday, OpenTime: domain.TimeOfDay{Hour: 7, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 14, Minute: 0}, IsOpen: true},
			{Day: domain.Sunday, OpenTime: domain.TimeOfDay{Hour: 8, Minute: 0}, CloseTime: domain.TimeOfDay{Hour: 12, Minute: 0}, IsOpen: true},
		},
		MinDelivery: 1000, CreatedAt: now,
	})

	log.Println("  ✓ Bakeries: La Boulangerie du Coin, Mie & Beurre, Le Fournil de Max")

	// ===================== PRODUCTS =====================
	// Bakery 1
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-1-1", BakeryID: "bakery-1", Name: "Croissant", Description: "Buttery, flaky classic", Price: 150, Category: "Viennoiseries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1555507036-ab1f4038808a?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-1-2", BakeryID: "bakery-1", Name: "Pain au Chocolat", Description: "Two bars of dark chocolate inside golden pastry", Price: 180, Category: "Viennoiseries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1623334044303-241021148842?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-1-3", BakeryID: "bakery-1", Name: "Baguette Tradition", Description: "Crusty sourdough baguette, 250g", Price: 130, Category: "Breads", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1608198093002-ad4e005484ec?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-1-4", BakeryID: "bakery-1", Name: "Pain de Campagne", Description: "Rustic country loaf with rye flour", Price: 420, Category: "Breads", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1589367920969-ab8e050bbb04?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-1-5", BakeryID: "bakery-1", Name: "Tarte aux Pommes", Description: "Apple tart with almond cream", Price: 450, Category: "Pastries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1562007908-17c67e878c88?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-1-6", BakeryID: "bakery-1", Name: "Éclair au Café", Description: "Coffee-filled choux pastry with chocolate glaze", Price: 380, Category: "Pastries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1525059696034-4967a8e1dca2?w=400"})

	// Bakery 2
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-2-1", BakeryID: "bakery-2", Name: "Sourdough Boule", Description: "48-hour fermented whole wheat sourdough", Price: 550, Category: "Breads", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1586444248902-2f64eddc13df?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-2-2", BakeryID: "bakery-2", Name: "Croissant au Beurre", Description: "All-butter croissant, 72-layer lamination", Price: 200, Category: "Viennoiseries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1555507036-ab1f4038808a?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-2-3", BakeryID: "bakery-2", Name: "Brioche Feuilletée", Description: "Buttery layered brioche, lightly sweet", Price: 350, Category: "Viennoiseries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1620921568790-8adcfb05e7a4?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-2-4", BakeryID: "bakery-2", Name: "Fougasse aux Olives", Description: "Provençal flatbread with black olives and herbs", Price: 380, Category: "Breads", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1509440159596-0249088772ff?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-2-5", BakeryID: "bakery-2", Name: "Tarte au Citron", Description: "Lemon meringue tart with shortcrust base", Price: 500, Category: "Pastries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1519915028121-7d3463d20b13?w=400"})

	// Bakery 3
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-3-1", BakeryID: "bakery-3", Name: "Pain au Levain", Description: "Wood-fired sourdough with thick crust", Price: 480, Category: "Breads", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1589367920969-ab8e050bbb04?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-3-2", BakeryID: "bakery-3", Name: "Kouign-Amann", Description: "Caramelized butter pastry from Brittany", Price: 320, Category: "Pastries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1509365390695-33aee754301f?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-3-3", BakeryID: "bakery-3", Name: "Pain aux Noix", Description: "Walnut bread with honey glaze", Price: 520, Category: "Breads", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1608198093002-ad4e005484ec?w=400"})
	bakeryRepo.SeedProduct(domain.Product{ID: "prod-3-4", BakeryID: "bakery-3", Name: "Chausson aux Pommes", Description: "Puff pastry turnover with apple compote", Price: 280, Category: "Viennoiseries", IsAvailable: true, PhotoURL: "https://images.unsplash.com/photo-1562007908-17c67e878c88?w=400"})

	log.Println("  ✓ Products: 15 across 3 bakeries")

	// ===================== ORDERS =====================
	order1 := &domain.Order{
		ID: "order-1", BakeryID: "bakery-1", UserID: "customer-1",
		Items: []domain.OrderItem{
			{ProductID: "prod-1-1", ProductName: "Croissant", Quantity: 3, UnitPrice: 150, Subtotal: 450},
			{ProductID: "prod-1-3", ProductName: "Baguette Tradition", Quantity: 1, UnitPrice: 130, Subtotal: 130},
		},
		ScheduledDay:  domain.Wednesday,
		ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 8, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 8, Minute: 30}},
		Status: domain.OrderStatusConfirmed, TotalAmount: 580, PaymentMethod: domain.PaymentMethodOnline,
		SelectionMode: domain.SelectionFixed, CreatedAt: now.Add(-2 * 24 * time.Hour), UpdatedAt: now.Add(-2 * 24 * time.Hour),
	}
	_ = orderRepo.Save(ctx, order1)

	order2 := &domain.Order{
		ID: "order-2", BakeryID: "bakery-2", UserID: "customer-1",
		Items: []domain.OrderItem{
			{ProductID: "prod-2-2", ProductName: "Croissant au Beurre", Quantity: 4, UnitPrice: 200, Subtotal: 800},
		},
		ScheduledDay:  domain.Friday,
		ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 7, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 7, Minute: 30}},
		Status: domain.OrderStatusPreparing, TotalAmount: 800, PaymentMethod: domain.PaymentMethodOnline,
		SelectionMode: domain.SelectionFixed, CreatedAt: now.Add(-1 * 24 * time.Hour), UpdatedAt: now.Add(-12 * time.Hour),
	}
	_ = orderRepo.Save(ctx, order2)

	order3 := &domain.Order{
		ID: "order-3", BakeryID: "bakery-1", UserID: "customer-2",
		Items: []domain.OrderItem{
			{ProductID: "prod-1-5", ProductName: "Tarte aux Pommes", Quantity: 1, UnitPrice: 450, Subtotal: 450},
			{ProductID: "prod-1-6", ProductName: "Éclair au Café", Quantity: 2, UnitPrice: 380, Subtotal: 760},
		},
		ScheduledDay:  domain.Saturday,
		ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 9, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 9, Minute: 30}},
		Status: domain.OrderStatusReady, TotalAmount: 1210, PaymentMethod: domain.PaymentMethodOnline,
		SelectionMode: domain.SelectionFixed, CreatedAt: now.Add(-3 * 24 * time.Hour), UpdatedAt: now.Add(-6 * time.Hour),
	}
	_ = orderRepo.Save(ctx, order3)

	log.Println("  ✓ Orders: 3 (confirmed, preparing, ready)")

	// ===================== RESERVATIONS =====================
	res1 := domain.Reservation{
		ID: "res-1", BakeryID: "bakery-2", UserID: "customer-2",
		Items: []domain.OrderItem{
			{ProductID: "prod-2-1", ProductName: "Sourdough Boule", Quantity: 1, UnitPrice: 550, Subtotal: 550},
			{ProductID: "prod-2-5", ProductName: "Tarte au Citron", Quantity: 1, UnitPrice: 500, Subtotal: 500},
		},
		ScheduledDay:  domain.Friday,
		ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 14, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 14, Minute: 30}},
		Status: domain.ReservationStatusConfirmed, TotalAmount: 1050, PaymentMethod: domain.PaymentMethodOnSpot,
		CreatedAt: now.Add(-1 * 24 * time.Hour),
	}
	_ = reservationRepo.Save(ctx, res1)

	res2 := domain.Reservation{
		ID: "res-2", BakeryID: "bakery-3", UserID: "customer-1",
		Items: []domain.OrderItem{
			{ProductID: "prod-3-2", ProductName: "Kouign-Amann", Quantity: 6, UnitPrice: 320, Subtotal: 1920},
		},
		ScheduledDay:  domain.Saturday,
		ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 10, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 10, Minute: 30}},
		Status: domain.ReservationStatusReady, TotalAmount: 1920, PaymentMethod: domain.PaymentMethodOnSpot,
		CreatedAt: now.Add(-2 * 24 * time.Hour),
	}
	_ = reservationRepo.Save(ctx, res2)

	log.Println("  ✓ Reservations: 2 (confirmed, ready)")

	// ===================== RECURRING ORDERS =====================
	recurring1 := &domain.RecurringOrder{
		ID: "recurring-1", UserID: "customer-1", BakeryID: "bakery-1",
		Items: []domain.OrderItem{
			{ProductID: "prod-1-1", ProductName: "Croissant", Quantity: 2, UnitPrice: 150, Subtotal: 300},
			{ProductID: "prod-1-3", ProductName: "Baguette Tradition", Quantity: 1, UnitPrice: 130, Subtotal: 130},
		},
		ScheduledDay:  domain.Monday,
		ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 7, Minute: 30}, EndTime: domain.TimeOfDay{Hour: 8, Minute: 0}},
		Frequency: domain.FrequencyWeekly, SelectionMode: domain.SelectionFixed, Active: true,
		CreatedAt: now.Add(-14 * 24 * time.Hour), UpdatedAt: now.Add(-14 * 24 * time.Hour),
	}
	_ = recurringRepo.Save(ctx, recurring1)

	recurring2 := &domain.RecurringOrder{
		ID: "recurring-2", UserID: "customer-2", BakeryID: "bakery-2",
		Items: []domain.OrderItem{
			{ProductID: "prod-2-2", ProductName: "Croissant au Beurre", Quantity: 4, UnitPrice: 200, Subtotal: 800},
		},
		ScheduledDay:  domain.Wednesday,
		ScheduledTime: domain.TimeSlot{StartTime: domain.TimeOfDay{Hour: 7, Minute: 0}, EndTime: domain.TimeOfDay{Hour: 7, Minute: 30}},
		Frequency: domain.FrequencyBiWeekly, SelectionMode: domain.SelectionRandomFavorites, Active: true,
		CreatedAt: now.Add(-7 * 24 * time.Hour), UpdatedAt: now.Add(-7 * 24 * time.Hour),
	}
	_ = recurringRepo.Save(ctx, recurring2)

	log.Println("  ✓ Recurring orders: 2 (weekly, bi-weekly)")

	// ===================== REGISTRATION TOKEN =====================
	demoToken := &domain.RegistrationToken{
		ID: "token-1", Token: "DEMO1234", Email: "newbaker@example.com",
		BakeryName: "Demo Bakery", ExpiresAt: now.Add(30 * 24 * time.Hour), Used: false, CreatedAt: now,
	}
	_ = tokenRepo.Save(ctx, demoToken)

	log.Println("  ✓ Registration token: DEMO1234 (for testing baker registration)")

	log.Println("=== Seed complete ===")
	log.Println("")
	log.Println("  Test accounts:")
	log.Println("  ┌────────────────┬──────────────┬─────────┐")
	log.Println("  │ Username       │ Password     │ Role    │")
	log.Println("  ├────────────────┼──────────────┼─────────┤")
	log.Println("  │ admin          │ admin123     │ Admin   │")
	log.Println("  │ baker_jean     │ baker123     │ Seller  │")
	log.Println("  │ baker_marie    │ baker123     │ Seller  │")
	log.Println("  │ alice          │ customer123  │ Customer│")
	log.Println("  │ bob            │ customer123  │ Customer│")
	log.Println("  └────────────────┴──────────────┴─────────┘")
	log.Println("  Baker registration code: DEMO1234")
	log.Println("")
}
