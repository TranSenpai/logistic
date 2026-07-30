package di

import (
	"os"

	"auth_service/internal/biz"
	"auth_service/internal/controller"
	"auth_service/internal/mapper/generated"
	"auth_service/internal/repo"

	entclient "auth_service/internal/common/ent_client"

	pb "github.com/logistic/api/logistic/auth_service/v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
)

func Injection(grpcServer *grpc.Server) error {
	clientDb, err := entclient.NewConnection()
	if err != nil {
		return err
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev_secret_change_before_production"
	}

	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	authMapper := &generated.AuthMapperImpl{}
	authRepo := repo.NewAuthRepo(clientDb, authMapper)
	authService := biz.NewAuthService(authRepo, jwtSecret, oauthConfig)
	controller := controller.NewGrpcHandler(authService, authMapper)
	pb.RegisterAuthServiceServer(grpcServer, controller)

	return nil
}
