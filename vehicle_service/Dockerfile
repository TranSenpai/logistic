FROM docker.io/library/golang:1.26.4-alpine3.24 AS build
WORKDIR /workspace
COPY go.work go.work.sum ./
COPY ./go.mod ./go.sum ./
COPY api/ ./api/
COPY pkg/ ./pkg/
COPY vehicle_service/ ./vehicle_service/
RUN rm -f go.work go.work.sum && go work init . ./api ./pkg ./vehicle_service
WORKDIR /workspace/vehicle_service
RUN go build -o vehicle_service ./cmd

FROM docker.io/library/alpine:3.24
WORKDIR /app
COPY --from=build /workspace/vehicle_service/vehicle_service .
CMD ["./vehicle_service"]
