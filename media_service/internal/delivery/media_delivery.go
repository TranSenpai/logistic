package http

import (
	"media_service/internal/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	storage storage.FileStorage
}

func NewMediaHandler(storage storage.FileStorage) *MediaHandler {
	return &MediaHandler{storage: storage}
}

// UploadFile godoc
// @Summary      Upload an image to Cloudinary
// @Description  Uploads a file via multipart/form-data and streams it to Cloudinary. Returns the public_id and url.
// @Tags         Media
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "The image file to upload"
// @Param        folder formData string false "The target folder in Cloudinary (e.g., logistics_images)"
// @Param        prefix formData string false "The prefix for the file name (e.g., img_)"
// @Param        Accept-Language header string false "Language for error messages (vi, en, jp)"
// @Success      200 {object} map[string]interface{} "message, file_name, public_id, url"
// @Failure      400 {object} map[string]interface{} "Missing parameters"
// @Failure      500 {object} map[string]interface{} "Internal server error or Cloudinary error"
// @Router       /upload [post]
func (h *MediaHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing file in format-data with key 'file'"})
		return
	}

	fileName, publicID, url, err := h.storage.Upload(c.Request.Context(), file, c.Request.FormValue("folder"), c.Request.FormValue("prefix"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Upload successfully!",
		"file_name": fileName,
		"public_id": publicID,
		"url":       url,
	})
}

// DeleteFile godoc
// @Summary      Delete an image from Cloudinary
// @Description  Deletes an image from Cloudinary using its public_id.
// @Tags         Media
// @Accept       json
// @Produce      json
// @Param        publicID path string true "The public_id of the image to delete"
// @Param        Accept-Language header string false "Language for error messages (vi, en, jp)"
// @Success      200 {object} map[string]interface{} "message: Delete successfully!"
// @Failure      400 {object} map[string]interface{} "Missing public_id"
// @Failure      500 {object} map[string]interface{} "Internal server error or Cloudinary error"
// @Router       /delete/{publicID} [delete]
func (h *MediaHandler) DeleteFile(c *gin.Context) {
	publicID := c.Param("publicID")
	if publicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing public_id in URL parameter"})
		return
	}

	err := h.storage.Delete(c.Request.Context(), publicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delete successfully!"})
}
