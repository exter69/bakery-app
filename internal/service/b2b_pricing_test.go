package service

import (
	"testing"

	"github.com/lucatorrekens/bakery-app/internal/domain"
)

func TestComputePricing_NoDiscount(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 2, UnitPrice: 500},
		{ProductID: "p2", Quantity: 3, UnitPrice: 1000},
	}
	result := computePricing(items, 0, 6)

	if result.SubtotalHT != 4000 {
		t.Errorf("SubtotalHT = %d, want 4000", result.SubtotalHT)
	}
	if result.DiscountAmount != 0 {
		t.Errorf("DiscountAmount = %d, want 0", result.DiscountAmount)
	}
	// TVA = 4000 * 6 / 100 = 240
	if result.TVAAmount != 240 {
		t.Errorf("TVAAmount = %d, want 240", result.TVAAmount)
	}
	if result.TotalTTC != 4240 {
		t.Errorf("TotalTTC = %d, want 4240", result.TotalTTC)
	}
}

func TestComputePricing_WithProDiscount(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 10, UnitPrice: 1000},
	}
	// 10 * 1000 = 10000 subtotal, 5% discount = 500, taxable = 9500, TVA 6% = 570
	result := computePricing(items, 5, 6)

	if result.SubtotalHT != 10000 {
		t.Errorf("SubtotalHT = %d, want 10000", result.SubtotalHT)
	}
	if result.DiscountAmount != 500 {
		t.Errorf("DiscountAmount = %d, want 500", result.DiscountAmount)
	}
	if result.TVAAmount != 570 {
		t.Errorf("TVAAmount = %d, want 570", result.TVAAmount)
	}
	if result.TotalTTC != 10070 {
		t.Errorf("TotalTTC = %d, want 10070", result.TotalTTC)
	}
}

func TestComputePricing_CustomVATRate(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: 10000},
	}
	// subtotal = 10000, no discount, TVA 21% = 2100
	result := computePricing(items, 0, 21)

	if result.TVARate != 21 {
		t.Errorf("TVARate = %d, want 21", result.TVARate)
	}
	if result.TVAAmount != 2100 {
		t.Errorf("TVAAmount = %d, want 2100", result.TVAAmount)
	}
	if result.TotalTTC != 12100 {
		t.Errorf("TotalTTC = %d, want 12100", result.TotalTTC)
	}
}

func TestComputeFullPricing_NoTiers(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 5, UnitPrice: 1000},
	}
	// subtotal = 5000, pro 1% = 50, vol 0%, taxable = 4950, TVA 6% = 297
	result := computeFullPricing(items, 1, 6, 0, nil)

	if result.SubtotalHT != 5000 {
		t.Errorf("SubtotalHT = %d, want 5000", result.SubtotalHT)
	}
	if result.ProDiscountRate != 1 {
		t.Errorf("ProDiscountRate = %d, want 1", result.ProDiscountRate)
	}
	if result.ProDiscountAmt != 50 {
		t.Errorf("ProDiscountAmt = %d, want 50", result.ProDiscountAmt)
	}
	if result.VolDiscountRate != 0 {
		t.Errorf("VolDiscountRate = %d, want 0", result.VolDiscountRate)
	}
	if result.VolDiscountAmt != 0 {
		t.Errorf("VolDiscountAmt = %d, want 0", result.VolDiscountAmt)
	}
	if result.TVAAmount != 297 {
		t.Errorf("TVAAmount = %d, want 297", result.TVAAmount)
	}
	if result.TotalTTC != 5247 {
		t.Errorf("TotalTTC = %d, want 5247", result.TotalTTC)
	}
}

func TestComputeFullPricing_WithVolumeTier(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 10, UnitPrice: 10000},
	}
	tiers := []domain.VolumeTier{
		{ID: "t1", MinMonthlySpend: 150000, DiscountPercent: 8},
		{ID: "t2", MinMonthlySpend: 200000, DiscountPercent: 10},
	}

	// Monthly spend 160000 means we qualify for tier 1 (8%)
	// subtotal = 100000, pro 1% = 1000, afterPro = 99000
	// vol 8% of 99000 = 7920, taxable = 91080
	// TVA 6% of 91080 = 5464 (truncated)
	// Total = 91080 + 5464 = 96544
	result := computeFullPricing(items, 1, 6, 160000, tiers)

	if result.ProDiscountAmt != 1000 {
		t.Errorf("ProDiscountAmt = %d, want 1000", result.ProDiscountAmt)
	}
	if result.VolDiscountRate != 8 {
		t.Errorf("VolDiscountRate = %d, want 8", result.VolDiscountRate)
	}
	if result.VolDiscountAmt != 7920 {
		t.Errorf("VolDiscountAmt = %d, want 7920", result.VolDiscountAmt)
	}
	if result.CurrentTier == nil || result.CurrentTier.ID != "t1" {
		t.Errorf("CurrentTier should be t1")
	}
	if result.NextTier == nil || result.NextTier.ID != "t2" {
		t.Errorf("NextTier should be t2")
	}
	if result.SpendToNextTier != 40000 {
		t.Errorf("SpendToNextTier = %d, want 40000", result.SpendToNextTier)
	}
	if result.TotalTTC != 96544 {
		t.Errorf("TotalTTC = %d, want 96544 (got %d)", 96544, result.TotalTTC)
	}
}

func TestComputeFullPricing_MaxTierReached(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: 50000},
	}
	tiers := []domain.VolumeTier{
		{ID: "t1", MinMonthlySpend: 150000, DiscountPercent: 8},
		{ID: "t2", MinMonthlySpend: 200000, DiscountPercent: 10},
	}

	// Monthly spend 250000 qualifies for max tier (10%)
	result := computeFullPricing(items, 1, 6, 250000, tiers)

	if result.VolDiscountRate != 10 {
		t.Errorf("VolDiscountRate = %d, want 10", result.VolDiscountRate)
	}
	if result.CurrentTier == nil || result.CurrentTier.ID != "t2" {
		t.Errorf("CurrentTier should be t2")
	}
	if result.NextTier != nil {
		t.Errorf("NextTier should be nil at max tier, got %+v", result.NextTier)
	}
	if result.SpendToNextTier != 0 {
		t.Errorf("SpendToNextTier = %d, want 0", result.SpendToNextTier)
	}
}

func TestComputeFullPricing_BelowAllTiers(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, UnitPrice: 10000},
	}
	tiers := []domain.VolumeTier{
		{ID: "t1", MinMonthlySpend: 150000, DiscountPercent: 8},
		{ID: "t2", MinMonthlySpend: 200000, DiscountPercent: 10},
	}

	// Monthly spend 50000 doesn't qualify for any tier
	result := computeFullPricing(items, 1, 6, 50000, tiers)

	if result.VolDiscountRate != 0 {
		t.Errorf("VolDiscountRate = %d, want 0", result.VolDiscountRate)
	}
	if result.VolDiscountAmt != 0 {
		t.Errorf("VolDiscountAmt = %d, want 0", result.VolDiscountAmt)
	}
	if result.CurrentTier != nil {
		t.Errorf("CurrentTier should be nil, got %+v", result.CurrentTier)
	}
	if result.NextTier == nil || result.NextTier.ID != "t1" {
		t.Errorf("NextTier should be t1")
	}
	if result.SpendToNextTier != 100000 {
		t.Errorf("SpendToNextTier = %d, want 100000", result.SpendToNextTier)
	}
}

func TestComputeFullPricing_EmptyItems(t *testing.T) {
	tiers := []domain.VolumeTier{
		{ID: "t1", MinMonthlySpend: 150000, DiscountPercent: 8},
	}

	result := computeFullPricing(nil, 1, 6, 0, tiers)

	if result.SubtotalHT != 0 {
		t.Errorf("SubtotalHT = %d, want 0", result.SubtotalHT)
	}
	if result.TotalTTC != 0 {
		t.Errorf("TotalTTC = %d, want 0", result.TotalTTC)
	}
}
