.PHONY: run build test swag

# Khởi chạy service với chế độ hot-reload
run:
	air

# Build file thực thi (binary)
build:
	go build -o ./tmp/main ./cmd

# Chạy Unit Test
test:
	go test ./... -v

# Tự động sinh tài liệu Swagger từ code (cho HTTP REST API)
swag:
	swag init -g cmd/main.go -d ./,../pkg
