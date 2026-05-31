package http

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, handler *MediaHandler) {
	v1 := r.Group("/api/v1/media")
	{
		v1.POST("/upload", handler.UploadFile)
		v1.DELETE("/delete/:publicID", handler.DeleteFile)
	}
}
