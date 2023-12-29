FROM alpine:latest
ENTRYPOINT ["/go-serve-s3"]
COPY go-serve-s3 /
