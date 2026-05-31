package main

import (
	"log"
	"os"

	"goBackend/pkg/i18n"

	"github.com/joho/godotenv"
)

// @title 				Media Service API
// @version 			1.0
// @description 		Microservice xử lý file và ảnh cho Logistics OS.
// @host 				localhost:8082
// @BasePath 			/api/v1/mediaS
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Cảnh báo: Không tìm thấy file .env")
	}

	i18n.InitI18n("../pkg/i18n/locales")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	app, err := NewApp(port)
	if err != nil {
		log.Fatalf("Lỗi khởi tạo App: %v", err)
	}

	log.Printf("==== Media Service đang chạy tại cổng %s ====", port)
	if err := app.Start(); err != nil {
		log.Fatalf("Lỗi khởi động server: %v", err)
	}
}
