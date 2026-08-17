FROM docker.io/library/alpine:3.24
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY user_service_bin .
RUN chmod +x user_service_bin
CMD ["./user_service_bin"]
