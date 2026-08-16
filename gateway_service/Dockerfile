FROM docker.io/library/alpine:3.24
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY gateway_service_bin .
RUN chmod +x gateway_service_bin
CMD ["./gateway_service_bin"]