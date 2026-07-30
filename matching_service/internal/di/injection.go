package di

import (
	"matching_service/internal/biz"
	"matching_service/internal/controller"
	"matching_service/internal/repo"

	"google.golang.org/grpc"

	entclient "matching_service/internal/common/ent_client"

	pb "github.com/logistic/api/logistic/matching_service/v1"
)

func Injection(grpcServer *grpc.Server) error {
	client, err := entclient.NewConnection()
	if err != nil {
		return err
	}

	repo := repo.NewMatchingRepo(client)
	engine := biz.NewGeoHashEngine()
	biz := biz.NewMatchingEngine(repo, engine)

	controller := controller.NewMatchingController(biz)
	pb.RegisterMatchingEngineServiceServer(grpcServer, controller)

	return nil
}
