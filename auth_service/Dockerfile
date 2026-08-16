FROM docker.io/library/alpine:3.24
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY auth_service_bin .
RUN chmod +x auth_service_bin
CMD ["./auth_service_bin"]
