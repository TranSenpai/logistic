package di

import (
	"fmt"
	http "media_service/internal/delivery"
	"media_service/internal/storage/cloudinary"
	"os"

	cld "github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
)

func Injection(r *gin.Engine) error {
	url := os.Getenv("CLOUDINARY_URL")
	if url == "" {
		return fmt.Errorf("missing CLOUDINARY_URL in environment")
	}

	cldClient, err := cld.NewFromURL(url)
	if err != nil {
		return fmt.Errorf("failed to connect to cloudinary: %w", err)
	}

	cloudStorage := cloudinary.NewCloudinaryStorage(cldClient)
	mediaHandler := http.NewMediaHandler(cloudStorage)

	http.SetupRoutes(r, mediaHandler)

	return nil
}
