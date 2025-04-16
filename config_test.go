package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigFromEnv(t *testing.T) {
	t.Setenv("APP_SERVER_HOST", "127.0.0.1")
	t.Setenv("APP_SERVER_PORT", "3000")
	t.Setenv("APP_S3_BUCKET", "test-bucket")
	t.Setenv("APP_S3_REGION", "us-west-1")
	t.Setenv("APP_S3_ENDPOINT_URL", "http://127.0.0.1:9090")
	t.Setenv("APP_S3_USE_PATH_STYLE", "true")
	t.Setenv("APP_CACHING_CAPACITY_ITEMS", "512")
	t.Setenv("APP_CACHING_CAPACITY_BYTES", "26214400")
	t.Setenv("APP_CACHING_TTL", "42m42s")

	actual := NewConfigFromEnv()

	assert.Equal(t, Config{
		ServerHost:           "127.0.0.1",
		ServerPort:           3000,
		S3Bucket:             "test-bucket",
		S3Region:             "us-west-1",
		S3EndpointURL:        "http://127.0.0.1:9090",
		S3UsePathStyle:       true,
		CachingCapacityItems: 512,
		CachingCapacityBytes: 25 * 1024 * 1024,
		CachingTTL:           42*time.Minute + 42*time.Second,
	}, actual)
}

func TestNewConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("APP_S3_BUCKET", "test-bucket")

	cfg := NewConfigFromEnv()

	assert.Equal(t, "0.0.0.0", cfg.ServerHost)
	assert.Equal(t, uint16(8080), cfg.ServerPort)
	assert.Equal(t, 1024, cfg.CachingCapacityItems)
	assert.Equal(t, 50*1024*1024, cfg.CachingCapacityBytes)
	assert.Equal(t, 10*time.Minute, cfg.CachingTTL)
}

func TestNewConfigFromEnv_Errors(t *testing.T) {
	t.Run("missing required field", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() { NewConfigFromEnv() })
	})

	t.Run("invalid server port", func(t *testing.T) {
		t.Setenv("APP_S3_BUCKET", "test-bucket")
		t.Setenv("APP_SERVER_PORT", "invalid")
		assert.Panics(t, func() { NewConfigFromEnv() })
	})

	t.Run("invalid use path style", func(t *testing.T) {
		t.Setenv("APP_S3_BUCKET", "test-bucket")
		t.Setenv("APP_S3_USE_PATH_STYLE", "invalid")
		assert.Panics(t, func() { NewConfigFromEnv() })
	})

	t.Run("invalid caching capacity items", func(t *testing.T) {
		t.Setenv("APP_S3_BUCKET", "test-bucket")
		t.Setenv("APP_CACHING_CAPACITY_ITEMS", "invalid")
		assert.Panics(t, func() { NewConfigFromEnv() })
	})

	t.Run("invalid caching capacity bytes", func(t *testing.T) {
		t.Setenv("APP_S3_BUCKET", "test-bucket")
		t.Setenv("APP_CACHING_CAPACITY_BYTES", "invalid")
		assert.Panics(t, func() { NewConfigFromEnv() })
	})

	t.Run("invalid caching TTL", func(t *testing.T) {
		t.Setenv("APP_S3_BUCKET", "test-bucket")
		t.Setenv("APP_CACHING_TTL", "invalid")
		assert.Panics(t, func() { NewConfigFromEnv() })
	})
}
