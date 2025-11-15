FROM alpine:latest
RUN apk add --no-cache --update ca-certificates tzdata
ENTRYPOINT ["/go-serve-s3"]
COPY go-serve-s3 /
