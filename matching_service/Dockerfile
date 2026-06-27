FROM docker.io/library/golang:1.26.4-alpine3.24 AS build

WORKDIR /workspace

COPY go.work go.work.sum ./
COPY ./go.mod ./go.sum ./
# MẸO: Xóa các service không copy khỏi file go.work để Go không báo lỗi
RUN go work edit -dropuse=./auth_service && \
    go work edit -dropuse=./media_service

# CHỈ copy những thư mục thật sự cần thiết cho matching_service
COPY api/ ./api/
COPY pkg/ ./pkg/
COPY matching_service/ ./matching_service/

# Di chuyển vào thư mục của matching_service để build
WORKDIR /workspace/matching_service
RUN go build -o matching_service ./cmd

FROM docker.io/library/alpine:3.24
WORKDIR /app
COPY --from=build /workspace/matching_service/matching_service .
CMD ["./matching_service"]
