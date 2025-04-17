# go-serve-s3

[![Build Status](https://github.com/yegor-usoltsev/go-serve-s3/actions/workflows/ci.yml/badge.svg)](https://github.com/yegor-usoltsev/go-serve-s3/actions)
[![Codecov](https://codecov.io/gh/yegor-usoltsev/go-serve-s3/graph/badge.svg?token=5I7K9PUI0P)](https://codecov.io/gh/yegor-usoltsev/go-serve-s3)
[![GitHub Release](https://img.shields.io/github/v/release/yegor-usoltsev/go-serve-s3?sort=semver)](https://github.com/yegor-usoltsev/go-serve-s3/releases)
[![Docker Image (docker.io)](https://img.shields.io/docker/v/yusoltsev/go-serve-s3?label=docker.io&sort=semver)](https://hub.docker.com/r/yusoltsev/go-serve-s3)
[![Docker Image (ghcr.io)](https://img.shields.io/docker/v/yusoltsev/go-serve-s3?label=ghcr.io&sort=semver)](https://github.com/yegor-usoltsev/go-serve-s3/pkgs/container/go-serve-s3)
[![Docker Image Size](https://img.shields.io/docker/image-size/yusoltsev/go-serve-s3?sort=semver&arch=amd64)](https://hub.docker.com/r/yusoltsev/go-serve-s3/tags)

A compact tool for serving static files from AWS S3 object storage with in-memory caching.

## Usage

### AWS S3

Minimal example using AWS S3 (using credentials from host's `~/.aws` directory):

```bash
docker run -d \
  -e APP_S3_BUCKET=my-bucket \
  -v ~/.aws:/root/.aws:ro \
  -p 8080:8080 \
  yusoltsev/go-serve-s3:latest
```

### MinIO

Minimal example using MinIO (local S3-compatible storage):

```bash
docker run -d \
  -e APP_S3_BUCKET=my-bucket \
  -e APP_S3_REGION=us-east-1 \
  -e APP_S3_ENDPOINT_URL=http://minio:9000 \
  -e APP_S3_USE_PATH_STYLE=true \
  -e AWS_ACCESS_KEY_ID=minioadmin \
  -e AWS_SECRET_ACCESS_KEY=minioadmin \
  -p 8080:8080 \
  yusoltsev/go-serve-s3:latest
```

### Environment Variables

| KEY                          | TYPE       | DEFAULT             | REQUIRED |
| ---------------------------- | ---------- | ------------------- | -------- |
| `APP_SERVER_HOST`            | `string`   | `0.0.0.0`           | Yes      |
| `APP_SERVER_PORT`            | `uint16`   | `8080`              | Yes      |
| `APP_S3_BUCKET`              | `string`   |                     | Yes      |
| `APP_S3_REGION`              | `string`   |                     | No       |
| `APP_S3_ENDPOINT_URL`        | `string`   |                     | No       |
| `APP_S3_USE_PATH_STYLE`      | `bool`     |                     | No       |
| `APP_CACHING_CAPACITY_ITEMS` | `int`      | `1024`              | Yes      |
| `APP_CACHING_CAPACITY_BYTES` | `int`      | `52428800` (50 MiB) | Yes      |
| `APP_CACHING_TTL`            | `Duration` | `10m` (10 minutes)  | Yes      |

You should also provide valid AWS credentials using `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, or through other
supported environment variables. For details, refer to
the [AWS SDK documentation](https://docs.aws.amazon.com/sdkref/latest/guide/environment-variables.html).

## Docker Images

This application is delivered as a multi-platform Docker image and is available for download from two image registries
of choice: [yusoltsev/go-serve-s3](https://hub.docker.com/r/yusoltsev/go-serve-s3)
and [ghcr.io/yegor-usoltsev/go-serve-s3](https://github.com/yegor-usoltsev/go-serve-s3/pkgs/container/go-serve-s3).

## Versioning

This project uses [Semantic Versioning](https://semver.org)

## Contributing

Pull requests are welcome. For major changes,
please [open an issue](https://github.com/yegor-usoltsev/go-serve-s3/issues/new) first to discuss what you would
like to change. Please make sure to update tests as appropriate.

## License

[MIT](https://github.com/yegor-usoltsev/go-serve-s3/blob/main/LICENSE)
