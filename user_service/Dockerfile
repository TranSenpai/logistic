FROM docker.io/library/golang:1.26.4-alpine3.24 AS build
WORKDIR /workspace
COPY go.work go.work.sum ./
COPY ./go.mod ./go.sum ./
COPY api/ ./api/
COPY pkg/ ./pkg/
COPY user_service/ ./user_service/
RUN rm -f go.work go.work.sum && go work init . ./api ./pkg ./user_service
WORKDIR /workspace/user_service
RUN go build -o user_service ./cmd

FROM docker.io/library/alpine:3.24
WORKDIR /app
COPY --from=build /workspace/user_service/user_service .
CMD ["./user_service"]
