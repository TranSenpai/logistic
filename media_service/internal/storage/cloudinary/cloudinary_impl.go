package cloudinary

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/cloudinary/cloudinary-go/v2"
)

type cloudinaryStorage struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryStorage(client *cloudinary.Cloudinary) *cloudinaryStorage {
	return &cloudinaryStorage{client: client}
}

func (c *cloudinaryStorage) Upload(ctx context.Context, fileHeader *multipart.FileHeader, folder string, prefix string) (string, string, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", "", "", err
	}
	defer file.Close()

	fileName := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := c.client.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID: fileName,
		Folder:   folder,
	})

	if err != nil {
		return "", "", "", fmt.Errorf("gọi API Cloudinary thất bại: %w", err)
	}

	if result.Error.Message != "" {
		log.Printf("Cloudinary Upload Error: %s", result.Error.Message)
		return "", "", "", fmt.Errorf("cloudinary trả về lỗi: %s", result.Error.Message)
	}

	return fileName, result.PublicID, result.SecureURL, nil
}

func (c *cloudinaryStorage) Delete(ctx context.Context, publicID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("gọi API Cloudinary thất bại: %w", err)
	}

	log.Printf("Cloudinary Destroy Result: %s", result.Result)

	if result.Error.Message != "" {
		log.Printf("Cloudinary Destroy Error Detail: %s", result.Error.Message)
		return fmt.Errorf("cloudinary xóa file thất bại: %s", result.Error.Message)
	}

	return nil
}
