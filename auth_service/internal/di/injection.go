package di

import (
	"auth_service/internal/biz"
	"auth_service/internal/conf"
	"auth_service/internal/controller"
	"auth_service/internal/mapper/generated"
	"auth_service/internal/repo"

	entclient "auth_service/internal/common/ent_client"

	pb "github.com/logistic/api/logistic/auth_service/v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
)

func Injection(grpcServer *grpc.Server, cfg *conf.Config) error {
	clientDb, err := entclient.NewConnection(cfg.Database)
	if err != nil {
		return err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	authMapper := &generated.AuthMapperImpl{}
	authRepo := repo.NewAuthRepo(clientDb, authMapper)
	authService := biz.NewAuthService(authRepo, cfg.JWT.Secret, oauthConfig)
	controller := controller.NewGrpcHandler(authService, authMapper)
	pb.RegisterAuthServiceServer(grpcServer, controller)

	return nil
}
