FROM docker.io/library/alpine:3.24
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY vehicle_service_bin .
RUN chmod +x vehicle_service_bin
CMD ["./vehicle_service_bin"]
