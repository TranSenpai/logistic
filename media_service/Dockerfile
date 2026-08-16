FROM docker.io/library/alpine:3.24
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY media_service_bin .
RUN chmod +x media_service_bin
CMD ["./media_service_bin"]
