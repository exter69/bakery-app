package service

import (
	"context"
	"log"
	"time"

	"github.com/lucatorrekens/bakery-app/internal/domain"
	"github.com/lucatorrekens/bakery-app/internal/ws"
)

// StartExpirationWorker starts a background goroutine that periodically
// expires overdue bundles and releases overdue reservations.
// It runs every 60 seconds and stops when ctx is cancelled.
func StartExpirationWorker(ctx context.Context, svc domain.BundleService, hub *ws.Hub) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	log.Println("[BUNDLE-EXPIRY] worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[BUNDLE-EXPIRY] worker stopped")
			return
		case <-ticker.C:
			expired, err := svc.ExpireOverdueBundles(ctx)
			if err != nil {
				log.Printf("[BUNDLE-EXPIRY] error expiring bundles: %v", err)
			}
			released, err := svc.ReleaseOverdueReservations(ctx)
			if err != nil {
				log.Printf("[BUNDLE-EXPIRY] error releasing reservations: %v", err)
			}
			if expired > 0 || released > 0 {
				log.Printf("[BUNDLE-EXPIRY] expired=%d bundles, released=%d reservations", expired, released)
			}
		}
	}
}
