package biz

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"matching_service/internal/biz"
// 	entclient "matching_service/internal/common/ent_client"
// 	"matching_service/internal/entity"
// )

// func init() {
// 	// Chuyển working directory về thư mục gốc của matching_service
// 	// Để hàm entclient.NewConnection() có thể đọc được file "configs/.env"
// 	_ = os.Chdir("../..")
// }

// func setupRealRepo() biz.MatchingRepo {
// 	client, err := entclient.NewConnection()
// 	if err != nil {
// 		panic(fmt.Sprintf("failed connecting to postgres via entclient: %v", err))
// 	}

// 	if err := client.Schema.Create(context.Background()); err != nil {
// 		panic(fmt.Sprintf("failed creating schema resources: %v", err))
// 	}

// 	return repo.NewMatchingRepo(client)
// }

// // BenchmarkMatchingEngine tests the throughput of the engine hitting real DB.
// // Run with: go test -bench=. ./internal/biz
// func BenchmarkMatchingEngine(b *testing.B) {
// 	realRepo := setupRealRepo()
// 	spatialEngine := NewGeoHashEngine()
// 	engine := NewMatchingEngine(realRepo, spatialEngine)
// 	ctx := context.Background()

// 	// Pre-fill with Asks (Supply)
// 	for i := 0; i < 50; i++ {
// 		_ = engine.SubmitAsk(ctx, &entity.Ask{
// 			ID:                fmt.Sprintf("%d", i+1),
// 			DriverID:          fmt.Sprintf("%d", i+1),
// 			AvailableVolumeM3: 10,
// 			AvailableWeightKg: 1000,
// 			MinPrice:          50,
// 			CurrentLocation:   entity.Location{ZoneID: "HCM-Q1", Latitude: 10.762622, Longitude: 106.660172},
// 		})
// 	}

// 	b.ResetTimer() // Start measuring performance here

// 	for i := 0; i < b.N; i++ {
// 		_ = engine.SubmitBid(ctx, &entity.Bid{
// 			ID:          fmt.Sprintf("%d", i+1),
// 			UserID:      fmt.Sprintf("%d", i+1),
// 			VolumeM3:    5,
// 			WeightKg:    500,
// 			MaxPrice:    60,
// 			Origin:      entity.Location{ZoneID: "HCM-Q1", Latitude: 10.762622, Longitude: 106.660172},
// 			Destination: entity.Location{ZoneID: "HCM-Q1", Latitude: 10.762622, Longitude: 106.660172},
// 		})
// 	}
// }

// // TestMatchingEngine_BasicMatch verifies the core business logic with real DB
// func TestMatchingEngine_BasicMatch(t *testing.T) {
// 	realRepo := setupRealRepo()
// 	spatialEngine := NewGeoHashEngine()
// 	engine := NewMatchingEngine(realRepo, spatialEngine)
// 	ctx := context.Background()

// 	ask := &entity.Ask{
// 		ID:                "1",
// 		DriverID:          "1",
// 		AvailableVolumeM3: 10,
// 		AvailableWeightKg: 1000,
// 		MinPrice:          50,
// 		CurrentLocation:   entity.Location{ZoneID: "HCM-Q1", Latitude: 10.762622, Longitude: 106.660172},
// 	}
// 	err := engine.SubmitAsk(ctx, ask)
// 	if err != nil {
// 		t.Fatalf("failed to submit ask: %v", err)
// 	}

// 	bid := &entity.Bid{
// 		ID:          "1",
// 		UserID:      "1",
// 		VolumeM3:    5,
// 		WeightKg:    500,
// 		MaxPrice:    60,
// 		Origin:      entity.Location{ZoneID: "HCM-Q1", Latitude: 10.762622, Longitude: 106.660172},
// 		Destination: entity.Location{ZoneID: "HCM-Q1", Latitude: 10.762622, Longitude: 106.660172},
// 	}
// 	err = engine.SubmitBid(ctx, bid)
// 	if err != nil {
// 		t.Fatalf("failed to submit bid: %v", err)
// 	}

// 	t.Log("Basic match test passed with real DB integration!")
// 	// (Ghi chú: Lắng nghe MatchStream bị loại bỏ tạm vì code mới nhất trong matching_engine.go không còn đẩy vào channel)
// }
