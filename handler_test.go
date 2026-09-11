package main

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	siloImage     = "docker.io/pgsty/silo:RELEASE.2026-09-03T13-18-01Z"
	siloUser      = "minioadmin"
	siloPassword  = "minioadmin"
	bucketName    = "test-bucket"
	region        = "us-east-1"
	objectName    = "lorem-ipsum.txt"
	objectContent = "Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n"
)

func setupSilo(t *testing.T) *s3.Client {
	t.Helper()
	container, err := tc.Run(
		t.Context(),
		siloImage,
		tc.WithExposedPorts("9000/tcp"),
		tc.WithEnv(map[string]string{
			"MINIO_ROOT_USER":     siloUser,
			"MINIO_ROOT_PASSWORD": siloPassword,
		}),
		tc.WithCmd("server", "/data"),
		tc.WithWaitStrategy(wait.ForHTTP("/minio/health/live").WithPort("9000/tcp")),
	)
	require.NoError(t, err)
	tc.CleanupContainer(t, container)

	endpoint, err := container.PortEndpoint(t.Context(), "9000/tcp", "http")
	require.NoError(t, err)

	awsCfg, err := awsConfig.LoadDefaultConfig(
		t.Context(),
		awsConfig.WithRegion(region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(siloUser, siloPassword, "")),
	)
	require.NoError(t, err)
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})

	_, err = client.CreateBucket(t.Context(), &s3.CreateBucketInput{Bucket: aws.String(bucketName)})
	require.NoError(t, err)

	_, err = client.PutObject(t.Context(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
		Body:   strings.NewReader(objectContent),
	})
	require.NoError(t, err)

	t.Setenv("APP_S3_BUCKET", bucketName)
	t.Setenv("APP_S3_REGION", region)
	t.Setenv("APP_S3_ENDPOINT_URL", endpoint)
	t.Setenv("APP_S3_USE_PATH_STYLE", "true")
	t.Setenv("AWS_ACCESS_KEY_ID", siloUser)
	t.Setenv("AWS_SECRET_ACCESS_KEY", siloPassword)

	return client
}

func TestNewHandler(t *testing.T) {
	setupSilo(t)
	cfg, err := NewConfigFromEnv()
	require.NoError(t, err)

	serverHandler, err := NewHandler(cfg)
	require.NoError(t, err)
	handler := serverHandler.ServeHTTP

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
	client := setupSilo(t)
	cfg, err := NewConfigFromEnv()
	require.NoError(t, err)

	s3HTTPHandler, err := s3Handler(cfg)
	require.NoError(t, err)
	handler := s3HTTPHandler.ServeHTTP

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
		_, err := client.PutObject(t.Context(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(bigObjectName),
			Body:   strings.NewReader(bigObjectContent),
		})
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
		_, err := s3Handler(cfg)
		assert.Error(t, err)
	})

	t.Run("invalid caching capacity bytes", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			CachingCapacityItems: 1024,
			CachingCapacityBytes: -1,
		}
		_, err := s3Handler(cfg)
		assert.Error(t, err)
	})

	t.Run("invalid caching TTL", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			CachingCapacityItems: 1024,
			CachingCapacityBytes: 50 * 1024 * 1024,
			CachingTTL:           0,
		}
		_, err := s3Handler(cfg)
		assert.Error(t, err)
	})

	t.Run("invalid AWS config", func(t *testing.T) {
		t.Setenv("AWS_PROFILE", "non-existent")
		cfg := Config{
			CachingCapacityItems: 1024,
			CachingCapacityBytes: 50 * 1024 * 1024,
			CachingTTL:           10 * time.Minute,
		}
		_, err := s3Handler(cfg)
		assert.Error(t, err)
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
