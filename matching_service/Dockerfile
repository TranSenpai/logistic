FROM docker.io/library/alpine:3.24
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY matching_service_bin .
RUN chmod +x matching_service_bin
CMD ["./matching_service_bin"]
