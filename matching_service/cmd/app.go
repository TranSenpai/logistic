package main

import (
	"matching_service/internal/di"
	"net"

	"log"
	_ "matching_service/docs"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
)

type App struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewApp() (*App, error) {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	err = di.Injection(grpcServer)
	if err != nil {
		return nil, err
	}

	return &App{
		grpcServer: grpcServer,
		listener:   lis,
	}, nil
}

func (a *App) Start() error {
	// Start HTTP server for Swagger in a goroutine
	go func() {
		mux := http.NewServeMux()
		// Serve the raw swagger.json file
		mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "docs/logistics/v1/logistics.swagger.json")
		})

		// Serve the Swagger UI
		mux.HandleFunc("/swagger/", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8083/swagger.json"),
		))

		log.Println("Starting Swagger UI for Matching Service on :8083...")
		if err := http.ListenAndServe(":8083", mux); err != nil {
			log.Fatalf("Swagger server crashed: %v", err)
		}
	}()

	return a.grpcServer.Serve(a.listener) // Lắng nghe kết nối gRPC trên port 8081
}
