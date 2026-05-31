package di

import (
	"os"

	"auth_service/internal/biz"
	entclient "auth_service/internal/common/ent_client"
	"auth_service/internal/delivery"
	"auth_service/internal/repo"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func Injection(ginEngine *gin.Engine) error {
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
		Scopes:       []string{"openid", "email", "profile"}, // Lấy google ID, email và profile của user từ google
		Endpoint:     google.Endpoint,
	}

	authRepo := repo.NewAuthRepo(clientDb)
	authService := biz.NewAuthService(authRepo, jwtSecret, oauthConfig)

	httpHandler := delivery.NewHttpHandler(authService)
	httpHandler.RegisterRouter(ginEngine)

	return nil
}
