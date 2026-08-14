FROM docker.io/library/golang:1.26.4-alpine3.24 AS build

WORKDIR /workspace

COPY go.work go.work.sum ./
COPY ./go.mod ./go.sum ./
# CHỈ copy những thư mục thật sự cần thiết cho matching_service
COPY api/ ./api/
COPY pkg/ ./pkg/
COPY matching_service/ ./matching_service/

RUN rm -f go.work go.work.sum && go work init . ./matching_service

# Di chuyển vào thư mục của matching_service để build
WORKDIR /workspace/matching_service
RUN go build -o matching_service ./cmd

FROM docker.io/library/alpine:3.24
WORKDIR /app
COPY --from=build /workspace/matching_service/matching_service .
CMD ["./matching_service"]
