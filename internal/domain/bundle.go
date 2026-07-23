package domain

import (
	"fmt"
	"time"
)

// BundleType represents whether the bundle contents are specified or a surprise.
type BundleType string

const (
	BundleTypeCompose  BundleType = "compose"
	BundleTypeSurprise BundleType = "surprise"
)

// BundleStatus represents the lifecycle state of a surplus bundle.
type BundleStatus string

const (
	BundleStatusDraft     BundleStatus = "draft"
	BundleStatusPublished BundleStatus = "published"
	BundleStatusExpired   BundleStatus = "expired"
	BundleStatusSoldOut   BundleStatus = "sold_out"
)

// BundleReservationStatus represents the state of a bundle reservation.
type BundleReservationStatus string

const (
	BundleReservationPending   BundleReservationStatus = "pending"
	BundleReservationConfirmed BundleReservationStatus = "confirmed"
	BundleReservationPickedUp  BundleReservationStatus = "picked_up"
	BundleReservationReleased  BundleReservationStatus = "released"
	BundleReservationCancelled BundleReservationStatus = "cancelled"
)

// SurplusBundle represents a discounted end-of-day package.
type SurplusBundle struct {
	ID                string       `json:"id"`
	BakeryID          string       `json:"bakeryId"`
	Name              string       `json:"name"`
	Type              BundleType   `json:"type"`
	PhotoURL          string       `json:"photoUrl"`
	Description       string       `json:"description"`
	EstimatedValue    int64        `json:"estimatedValue"`
	OriginalPrice     int64        `json:"originalPrice"`
	DiscountedPrice   int64        `json:"discountedPrice"`
	QuantityTotal     int          `json:"quantityTotal"`
	QuantityRemaining int          `json:"quantityRemaining"`
	PickupStartTime   TimeOfDay    `json:"pickupStartTime"`
	PickupEndTime     TimeOfDay    `json:"pickupEndTime"`
	PublishedDate     string       `json:"publishedDate"`
	ExpiresAt         time.Time    `json:"expiresAt"`
	Status            BundleStatus `json:"status"`
	Items             []BundleItem `json:"items"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
}

// BundleItem represents a single item in a composé bundle.
type BundleItem struct {
	ID          string `json:"id"`
	BundleID    string `json:"bundleId"`
	ProductID   string `json:"productId,omitempty"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

// BundleReservation represents a customer's claim on a surplus bundle.
type BundleReservation struct {
	ID        string                  `json:"id"`
	BundleID  string                  `json:"bundleId"`
	UserID    string                  `json:"userId"`
	Status    BundleReservationStatus `json:"status"`
	CreatedAt time.Time               `json:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt"`
}

// BundleImpact holds community impact metrics.
type BundleImpact struct {
	TotalSaved    int     `json:"totalSaved"`
	WeightAvoided float64 `json:"weightAvoided"`
}

// ValidateBundle checks that a SurplusBundle satisfies all business rules.
func ValidateBundle(bundle SurplusBundle) error {
	if bundle.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(bundle.Name) > 100 {
		return fmt.Errorf("name must be at most 100 characters")
	}

	if bundle.Type != BundleTypeCompose && bundle.Type != BundleTypeSurprise {
		return fmt.Errorf("type must be %q or %q", BundleTypeCompose, BundleTypeSurprise)
	}

	if bundle.OriginalPrice <= 0 {
		return fmt.Errorf("original price must be greater than zero")
	}
	if bundle.DiscountedPrice <= 0 {
		return fmt.Errorf("discounted price must be greater than zero")
	}
	if bundle.DiscountedPrice >= bundle.OriginalPrice {
		return fmt.Errorf("discounted price must be less than original price")
	}

	if bundle.QuantityTotal < 1 {
		return fmt.Errorf("quantity total must be at least 1")
	}

	if !bundle.PickupStartTime.Before(bundle.PickupEndTime) {
		return fmt.Errorf("pickup start time must be before pickup end time")
	}

	// Type-specific validation
	if bundle.Type == BundleTypeCompose {
		if len(bundle.Items) == 0 {
			return fmt.Errorf("compose bundle must have at least one item")
		}
		for i, item := range bundle.Items {
			if item.Quantity < 1 {
				return fmt.Errorf("item %d quantity must be at least 1", i)
			}
		}
	}

	if bundle.Type == BundleTypeSurprise {
		if bundle.Description == "" {
			return fmt.Errorf("surprise bundle must have a description")
		}
		if len(bundle.Description) > 200 {
			return fmt.Errorf("description must be at most 200 characters")
		}
		if bundle.EstimatedValue <= 0 {
			return fmt.Errorf("surprise bundle estimated value must be greater than zero")
		}
	}

	return nil
}
