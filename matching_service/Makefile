.PHONY: run build test ent

# Khởi chạy service với chế độ hot-reload
run:
	air

# Build file thực thi (binary)
build:
	go build -o ./tmp/main ./cmd

# Chạy Unit Test
test:
	go test ./... -v

# Tự động sinh code Database ORM (Ent)
ent:
	go generate ./ent
