package di

import (
	"matching_service/internal/biz"
	entclient "matching_service/internal/common/ent_client"
	"matching_service/internal/delivery"
	"matching_service/internal/handler/grpchandler"
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

	// Create the handler (controller)
	handler := grpchandler.NewMatchingHandler(biz)

	// Let delivery layer register the handler to the gRPC Server
	delivery.RegisterGrpcRouter(grpcServer, handler)

	return nil
}
