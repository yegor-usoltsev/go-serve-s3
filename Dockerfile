FROM alpine:latest
RUN apk add --no-cache --update ca-certificates tini tzdata
ARG TARGETPLATFORM
ENTRYPOINT ["tini", "--"]
CMD ["/go-serve-s3"]
COPY $TARGETPLATFORM/go-serve-s3 /
