FROM alpine:latest
RUN apk add --no-cache --update ca-certificates tzdata
ARG TARGETPLATFORM
ENTRYPOINT ["/go-serve-s3"]
COPY $TARGETPLATFORM/go-serve-s3 /
