package di

import (
	"context"
	"log"

	"user_service/ent"
	"user_service/internal/biz"
	"user_service/internal/conf"
	"user_service/internal/controller"
	"user_service/internal/mapper/generated"
	"user_service/internal/repo"

	userv1 "github.com/logistic/api/logistic/user_service/v1"
	"google.golang.org/grpc"
)

func Injection(grpcServer *grpc.Server, cfg *conf.Config) (*ent.Client, error) {
	client, err := ent.Open(cfg.Database.Driver, cfg.Database.GetDataSource())
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
		return nil, err
	}

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
		return nil, err
	}

	userRepo := repo.NewUserRepo(client)
	appMapper := &generated.AppMapperImpl{}
	userEngine := biz.NewUserEngine(userRepo)
	userController := controller.NewUserController(userEngine, appMapper)

	userv1.RegisterUserServiceServer(grpcServer, userController)

	return client, nil
}
