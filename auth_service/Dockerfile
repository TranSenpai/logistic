# --- Stage 1: Build binary ---
FROM docker.io/library/golang:1.26.4-alpine3.24 AS build

WORKDIR /workspace

# Copy các file cấu hình module gốc và workspace
COPY go.work go.work.sum ./
COPY go.mod go.sum ./

# MẸO: Xóa các service không copy khỏi file go.work để Go không báo lỗi "no such file"
RUN go work edit -dropuse=./matching_service && \
    go work edit -dropuse=./media_service

# CHỈ copy những thư mục thật sự cần thiết cho auth_service
COPY api/ ./api/
COPY pkg/ ./pkg/
COPY auth_service/ ./auth_service/

# Di chuyển vào thư mục của auth_service để build
WORKDIR /workspace/auth_service

# Build file binary và xuất ra /app/auth_service
RUN go build -o /app/auth_service ./cmd

# --- Stage 2: Môi trường chạy siêu nhẹ ---
FROM docker.io/library/alpine:3.24

WORKDIR /app

# Copy file binary đã build từ stage 'build' sang
COPY --from=build /app/auth_service .

# Copy thư mục configs chứa file .env
COPY --from=build /workspace/auth_service/configs ./configs

# Chạy file binary
CMD ["./auth_service"]

# (Giải thích cờ -f: Vì chúng ta đang đứng ở thư mục gốc, Podman sẽ không biết Dockerfile nằm đâu, nên ta dùng -f để chỉ đường dẫn tới file Dockerfile)

# 61.52 MB
