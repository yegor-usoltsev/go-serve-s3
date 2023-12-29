FROM alpine:latest
COPY go-serve-s3 /go-serve-s3
ENTRYPOINT ["/go-serve-s3"]
