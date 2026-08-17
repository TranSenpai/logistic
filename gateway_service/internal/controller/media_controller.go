package controller

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/logistic/api/logistic/media_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MediaController struct {
	mediaClient pb.MediaServiceClient
}

func NewMediaController(mediaClient pb.MediaServiceClient) *MediaController {
	return &MediaController{
		mediaClient: mediaClient,
	}
}

// UploadFile godoc
// @Summary      Tải file lên hệ thống
// @Description  Nhận file qua multipart/form-data và gọi gRPC sang media_service để lưu trữ (ví dụ Cloudinary).
// @Tags         Media
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File cần tải lên"
// @Param        folder formData string false "Thư mục lưu trữ"
// @Param        prefix formData string false "Tiền tố tên file"
// @Success      200 {object} map[string]interface{} "Tải lên thành công, trả về URL"
// @Failure      400 {object} map[string]interface{} "Lỗi thiếu file hoặc file không hợp lệ"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/media/v1/upload [post]
func (c *MediaController) UploadFile(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": "file is required"})
		return
	}

	folder := ctx.PostForm("folder")
	prefix := ctx.PostForm("prefix")

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error", "message": "failed to open file"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error", "message": "failed to read file"})
		return
	}

	resp, err := c.mediaClient.UploadFile(ctx.Request.Context(), &pb.UploadFileRequest{
		FileContent: fileBytes,
		FileName:    fileHeader.Filename,
		Folder:      folder,
		Prefix:      prefix,
	})
	if err != nil {
		st, _ := status.FromError(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "upload_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"file_name": resp.FileName,
		"public_id": resp.PublicId,
		"url":       resp.Url,
		"message":   resp.Message,
	})
}

// DeleteFile godoc
// @Summary      Xóa file
// @Description  Xóa file trên hệ thống lưu trữ thông qua public_id.
// @Tags         Media
// @Produce      json
// @Param        publicID path string true "Public ID của file cần xóa"
// @Success      200 {object} map[string]interface{} "Xóa thành công"
// @Failure      400 {object} map[string]interface{} "Thiếu public_id"
// @Failure      404 {object} map[string]interface{} "Không tìm thấy file"
// @Failure      500 {object} map[string]interface{} "Lỗi server nội bộ"
// @Router       /api/media/v1/delete/{publicID} [delete]
func (c *MediaController) DeleteFile(ctx *gin.Context) {
	publicID := ctx.Param("publicID")
	if publicID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": "publicID is required"})
		return
	}

	resp, err := c.mediaClient.DeleteFile(ctx.Request.Context(), &pb.DeleteFileRequest{
		PublicId: publicID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": st.Message()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "delete_failed", "message": st.Message()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
	})
}
