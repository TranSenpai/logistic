package di

import (
	"context"
	"log"

	"vehicle_service/ent"
	"vehicle_service/internal/biz"
	"vehicle_service/internal/conf"
	"vehicle_service/internal/controller"
	"vehicle_service/internal/mapper/generated"
	"vehicle_service/internal/repo"

	vehiclev1 "github.com/logistic/api/logistic/vehicle_service/v1"
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

	vehicleRepo := repo.NewVehicleRepo(client)
	appMapper := &generated.AppMapperImpl{}
	vehicleEngine := biz.NewVehicleEngine(vehicleRepo)
	vehicleController := controller.NewVehicleController(vehicleEngine, appMapper)

	vehiclev1.RegisterVehicleServiceServer(grpcServer, vehicleController)

	return client, nil
}
