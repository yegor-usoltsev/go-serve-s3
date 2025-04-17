package main

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	tcMinio "github.com/testcontainers/testcontainers-go/modules/minio"
)

const (
	minioUser     = "minioadmin"
	minioPassword = "minioadmin"
	bucketName    = "test-bucket"
	region        = "us-east-1"
	objectName    = "lorem-ipsum.txt"
	objectContent = "Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n"
)

func setupMinio(t *testing.T) *minio.Client {
	t.Helper()
	container, err := tcMinio.Run(t.Context(), "minio/minio:latest")
	tc.CleanupContainer(t, container)
	require.NoError(t, err)

	endpoint, err := container.ConnectionString(t.Context())
	require.NoError(t, err)

	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(minioUser, minioPassword, "")})
	require.NoError(t, err)

	err = client.MakeBucket(t.Context(), bucketName, minio.MakeBucketOptions{Region: region})
	require.NoError(t, err)

	_, err = client.PutObject(t.Context(), bucketName, objectName, strings.NewReader(objectContent), -1, minio.PutObjectOptions{})
	require.NoError(t, err)

	t.Setenv("APP_S3_BUCKET", bucketName)
	t.Setenv("APP_S3_REGION", region)
	t.Setenv("APP_S3_ENDPOINT_URL", "http://"+endpoint)
	t.Setenv("APP_S3_USE_PATH_STYLE", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", minioUser)
	t.Setenv("AWS_SECRET_ACCESS_KEY", minioPassword)

	return client
}

func TestNewHandler(t *testing.T) { //nolint:paralleltest
	setupMinio(t)
	cfg := NewConfigFromEnv()

	handler := NewHandler(cfg).ServeHTTP

	assert.HTTPSuccess(t, handler, http.MethodGet, "/health", nil)
	assert.HTTPError(t, handler, http.MethodPost, "/health", nil)

	assert.HTTPSuccess(t, handler, http.MethodGet, "/", nil)
	assert.HTTPError(t, handler, http.MethodPost, "/", nil)

	assert.HTTPSuccess(t, handler, http.MethodGet, "/"+objectName, nil)
	assert.HTTPError(t, handler, http.MethodPost, "/"+objectName, nil)

	assert.HTTPError(t, handler, http.MethodGet, "/non-existent.txt", nil)
	assert.HTTPError(t, handler, http.MethodPost, "/non-existent.txt", nil)

	assert.HTTPRedirect(t, handler, http.MethodGet, "/../invalid.txt", nil)
	assert.HTTPRedirect(t, handler, http.MethodPost, "/../invalid.txt", nil)
}

func TestHealthHandler(t *testing.T) {
	t.Parallel()
	url := "/health"
	assert.HTTPStatusCode(t, healthHandler, http.MethodGet, url, nil, http.StatusOK)
	assert.HTTPBodyContains(t, healthHandler, http.MethodGet, url, nil, "OK")
}

func TestS3Handler(t *testing.T) {
	client := setupMinio(t)
	cfg := NewConfigFromEnv()

	handler := s3Handler(cfg).ServeHTTP

	t.Run("get root directory listing", func(t *testing.T) {
		t.Parallel()
		url := "/"
		assert.HTTPStatusCode(t, handler, http.MethodGet, url, nil, http.StatusOK)
		assert.HTTPBodyContains(t, handler, http.MethodGet, url, nil, objectName)
	})

	t.Run("get existing file", func(t *testing.T) {
		t.Parallel()
		url := "/" + objectName
		assert.HTTPStatusCode(t, handler, http.MethodGet, url, nil, http.StatusOK)
		assert.HTTPBodyContains(t, handler, http.MethodGet, url, nil, objectContent)
	})

	t.Run("get big file", func(t *testing.T) {
		t.Parallel()
		bigObjectName := "big-" + objectName
		bigObjectContent := func() string {
			var sb strings.Builder
			for i := range 100 {
				sb.WriteString(strconv.Itoa(i))
				sb.WriteRune(' ')
				sb.WriteString(objectContent)
			}
			return sb.String()
		}()
		_, err := client.PutObject(t.Context(), bucketName, bigObjectName, strings.NewReader(bigObjectContent), -1, minio.PutObjectOptions{})
		require.NoError(t, err)

		url := "/" + bigObjectName
		assert.HTTPStatusCode(t, handler, http.MethodGet, url, nil, http.StatusOK)
		assert.HTTPBodyContains(t, handler, http.MethodGet, url, nil, bigObjectContent)
	})

	t.Run("get non-existent file", func(t *testing.T) {
		t.Parallel()
		url := "/non-existent.txt"
		assert.HTTPStatusCode(t, handler, http.MethodGet, url, nil, http.StatusNotFound)
	})

	t.Run("get with invalid path", func(t *testing.T) {
		t.Parallel()
		url := "/../invalid.txt"
		assert.HTTPStatusCode(t, handler, http.MethodGet, url, nil, http.StatusNotFound)
	})
}

func TestS3Handler_Errors(t *testing.T) {
	t.Run("invalid caching capacity items", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			CachingCapacityItems: -1,
			CachingCapacityBytes: 50 * 1024 * 1024,
		}
		assert.Panics(t, func() { s3Handler(cfg) })
	})

	t.Run("invalid caching capacity bytes", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			CachingCapacityItems: 1024,
			CachingCapacityBytes: -1,
		}
		assert.Panics(t, func() { s3Handler(cfg) })
	})

	t.Run("invalid caching TTL", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			CachingCapacityItems: 1024,
			CachingCapacityBytes: 50 * 1024 * 1024,
			CachingTTL:           0,
		}
		assert.Panics(t, func() { s3Handler(cfg) })
	})

	t.Run("invalid AWS config", func(t *testing.T) {
		t.Setenv("AWS_PROFILE", "non-existent")
		cfg := Config{
			CachingCapacityItems: 1024,
			CachingCapacityBytes: 50 * 1024 * 1024,
			CachingTTL:           10 * time.Minute,
		}
		assert.Panics(t, func() { s3Handler(cfg) })
	})
}

func TestWithRecovery(t *testing.T) {
	t.Run("normal handler", func(t *testing.T) {
		t.Parallel()
		handler := withRecovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP
		assert.HTTPStatusCode(t, handler, http.MethodGet, "/", nil, http.StatusOK)
	})

	t.Run("panic handler", func(t *testing.T) {
		t.Parallel()
		handler := withRecovery(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic("test panic")
		})).ServeHTTP
		assert.HTTPStatusCode(t, handler, http.MethodGet, "/", nil, http.StatusInternalServerError)
	})
}
