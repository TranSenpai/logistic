package biz

import (
	"context"
	"fmt"
	"goBackend/internal/entity"
	"testing"
)

// BenchmarkMatchingEngine tests the throughput of the in-memory engine.
// Run with: go test -bench=. ./internal/biz
func BenchmarkMatchingEngine(b *testing.B) {
	engine := NewMatchingEngine()
	ctx := context.Background()

	// Pre-fill with Asks (Supply) to simulate a busy network
	for i := 0; i < 1000; i++ {
		_ = engine.SubmitAsk(ctx, &entity.Ask{
			ID:                fmt.Sprintf("ask-%d", i),
			AvailableVolumeM3: 10,
			AvailableWeightKg: 1000,
			MinPrice:          50,
			CurrentLocation:   entity.Location{ZoneID: "HCM-Q1"},
		})
	}

	b.ResetTimer() // Start measuring performance here

	for i := 0; i < b.N; i++ {
		// Submit Bids that will match the existing asks
		_ = engine.SubmitBid(ctx, &entity.Bid{
			ID:       fmt.Sprintf("bid-%d", i),
			VolumeM3: 5,
			WeightKg: 500,
			MaxPrice: 60,
			Origin:   entity.Location{ZoneID: "HCM-Q1"},
		})
	}
}

// TestMatchingEngine_BasicMatch verifies the core business logic
func TestMatchingEngine_BasicMatch(t *testing.T) {
	engine := NewMatchingEngine()
	ctx := context.Background()

	ask := &entity.Ask{
		ID:                "ask-1",
		AvailableVolumeM3: 10,
		AvailableWeightKg: 1000,
		MinPrice:          50,
		CurrentLocation:   entity.Location{ZoneID: "HCM-Q1"},
	}
	_ = engine.SubmitAsk(ctx, ask)

	bid := &entity.Bid{
		ID:       "bid-1",
		VolumeM3: 5,
		WeightKg: 500,
		MaxPrice: 60,
		Origin:   entity.Location{ZoneID: "HCM-Q1"},
	}
	_ = engine.SubmitBid(ctx, bid)

	// Consume channel to verify event driven pattern
	select {
	case match := <-engine.MatchStream():
		if match.AskID != "ask-1" || match.BidID != "bid-1" {
			t.Errorf("expected match ask-1 and bid-1, got %s and %s", match.AskID, match.BidID)
		}
		if match.Price != 50 {
			t.Errorf("expected settlement price to be 50, got %f", match.Price)
		}
	default:
		t.Errorf("expected a match event, got none")
	}
}
