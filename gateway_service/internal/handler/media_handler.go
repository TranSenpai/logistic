package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/logistic/api/logistic/media_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MediaHandler struct {
	mediaClient pb.MediaServiceClient
}

func NewMediaHandler(mediaClient pb.MediaServiceClient) *MediaHandler {
	return &MediaHandler{
		mediaClient: mediaClient,
	}
}

// UploadFile handles file upload
func (h *MediaHandler) UploadFile(ctx *gin.Context) {
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

	resp, err := h.mediaClient.UploadFile(ctx.Request.Context(), &pb.UploadFileRequest{
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

// DeleteFile handles file deletion
func (h *MediaHandler) DeleteFile(ctx *gin.Context) {
	publicID := ctx.Param("publicID")
	if publicID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "validation_failed", "message": "publicID is required"})
		return
	}

	resp, err := h.mediaClient.DeleteFile(ctx.Request.Context(), &pb.DeleteFileRequest{
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
