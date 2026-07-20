package di

import (
	logisticsv1 "goBackend/api/logistics/v1/gen/go/logistics/v1"
	"matching_service/internal/biz"
	entclient "matching_service/internal/common/ent_client"
	"matching_service/internal/delivery"
	"matching_service/internal/repo"

	"google.golang.org/grpc"
)

func Injection(grpcServer *grpc.Server) error {
	client, err := entclient.NewConnection()
	if err != nil {
		return err
	}

	repo := repo.NewMatchingRepo(client)
	engine := biz.NewGeoHashEngine()
	biz := biz.NewMatchingEngine(repo, engine)
	delivery := delivery.NewLogisticsDelivery(biz)
	logisticsv1.RegisterMatchingEngineServiceServer(grpcServer, delivery)

	return nil
}
