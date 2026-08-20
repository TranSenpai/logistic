// Package di ráp các tầng lại với nhau. Đây là NƠI DUY NHẤT biết mọi thứ:
// ent client, Redis, mapper, repo, biz, controller.
//
// Mọi tầng khác chỉ nhận dependency qua tham số hàm khởi tạo, nên đổi Postgres
// sang thứ khác hay bỏ Redis đều chỉ phải sửa file này.
package di

import (
	"context"
	"fmt"
	"log"

	"user_service/ent"
	"user_service/internal/biz"
	"user_service/internal/conf"
	"user_service/internal/controller"
	"user_service/internal/mapper/generated"
	"user_service/internal/repo"

	pb "github.com/logistic/api/logistic/user_service/v1"
	"github.com/logistic/pkg/cache"
	"google.golang.org/grpc"
)

// Container giữ các tài nguyên cần đóng lúc tắt service.
type Container struct {
	EntClient *ent.Client
	Cache     *cache.Client
}

func (c *Container) Close() {
	if c == nil {
		return
	}
	if c.Cache != nil {
		if err := c.Cache.Close(); err != nil {
			log.Printf("[user_service] closing redis failed: %v", err)
		}
	}
	if c.EntClient != nil {
		if err := c.EntClient.Close(); err != nil {
			log.Printf("[user_service] closing ent client failed: %v", err)
		}
	}
}

func Injection(grpcServer *grpc.Server, cfg *conf.Config) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("user_service: config is nil")
	}

	entClient, err := ent.Open(cfg.Database.Driver, cfg.Database.GetDataSource())
	if err != nil {
		return nil, fmt.Errorf("user_service: mở kết nối Postgres thất bại: %w", err)
	}

	if err := entClient.Schema.Create(context.Background()); err != nil {
		_ = entClient.Close()
		return nil, fmt.Errorf("user_service: tạo schema thất bại: %w", err)
	}

	// Redis hỏng KHÔNG được làm service chết. Ta log lại rồi chạy tiếp với
	// redisClient == nil; repo đã viết để chịu được trường hợp này.
	var redisClient *cache.Client
	if cfg.Redis.Enabled {
		redisClient, err = cache.New(cache.Config{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			Prefix:   cfg.Redis.Prefix,
		})
		if err != nil {
			log.Printf("[user_service] Redis không khả dụng (%v) — chạy tiếp KHÔNG cache", err)
			redisClient = nil
		} else {
			log.Printf("[user_service] Redis đã kết nối tại %s:%s (db=%d, prefix=%q)",
				cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB, cfg.Redis.Prefix)
		}
	} else {
		log.Printf("[user_service] Redis bị tắt bằng cấu hình — chạy KHÔNG cache")
	}

	appMapper := &generated.AppMapperImpl{}
	userRepo := repo.NewUserRepo(entClient, redisClient, appMapper)
	userEngine := biz.NewUserEngine(userRepo)
	userController := controller.NewUserController(userEngine, appMapper)

	pb.RegisterUserServiceServer(grpcServer, userController)

	return &Container{EntClient: entClient, Cache: redisClient}, nil
}
