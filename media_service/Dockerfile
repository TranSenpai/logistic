FROM docker.io/library/golang:1.26.4-alpine3.24 AS build
WORKDIR /workspace
COPY go.work go.work.sum ./
COPY ./go.mod ./go.sum ./
RUN go work edit -dropuse=./auth_service && \
    go work edit -dropuse=./matching_service
COPY api/ ./api/
COPY pkg/ ./pkg/
COPY media_service/ ./media_service/
WORKDIR /workspace/media_service
RUN go build -o media_service ./cmd

FROM docker.io/library/alpine:3.24
WORKDIR /app
COPY --from=build /workspace/media_service/media_service .
CMD ["./media_service"]
