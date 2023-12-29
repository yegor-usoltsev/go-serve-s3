# go-serve-s3

[![Build Status](https://github.com/yegor-usoltsev/go-serve-s3/actions/workflows/ci.yml/badge.svg)](https://github.com/yegor-usoltsev/go-serve-s3/actions)
[![GitHub Release](https://img.shields.io/github/v/release/yegor-usoltsev/go-serve-s3?sort=semver)](https://github.com/yegor-usoltsev/go-serve-s3/releases)
[![Docker Image (docker.io)](https://img.shields.io/docker/v/yusoltsev/go-serve-s3?label=docker.io&sort=semver)](https://hub.docker.com/r/yusoltsev/go-serve-s3)
[![Docker Image (ghcr.io)](https://img.shields.io/docker/v/yusoltsev/go-serve-s3?label=ghcr.io&sort=semver)](https://github.com/yegor-usoltsev/go-serve-s3/pkgs/container/go-serve-s3)

A compact tool for serving static files from AWS S3 object storage with in-memory caching.

### Environment Variables

| KEY               | TYPE      | DEFAULT   |
|-------------------|-----------|-----------|
| `APP_SERVER_HOST` | `String`  | `0.0.0.0` |
| `APP_SERVER_PORT` | `Integer` | `3000`    |
| `APP_S3_BUCKET`   | `String`  |           |
| `APP_S3_REGION`   | `String`  |           |

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
