package main

import (
	"context"
	"fmt"
	"net/http"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jszwec/s3fs/v2"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

type Handler struct {
	*http.ServeMux
}

func NewHandler(cfg Config) *Handler {
	h := &Handler{ServeMux: http.NewServeMux()}
	h.HandleFunc("GET /health", HealthHandler)
	h.Handle("GET /", NewS3Handler(cfg))
	return h
}

func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprint(w, "OK")
}

func NewS3Handler(cfg Config) http.Handler {
	memoryAdapter, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(cfg.CachingCapacityItems),
		memory.AdapterWithStorageCapacity(cfg.CachingCapacityBytes),
	)
	if err != nil {
		panic(err)
	}
	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memoryAdapter),
		cache.ClientWithTTL(cfg.CachingTTL),
		cache.ClientWithMethods([]string{http.MethodGet}),
		cache.ClientWithExpiresHeader(),
	)
	if err != nil {
		panic(err)
	}
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.DisableLogOutputChecksumValidationSkipped = true
		if cfg.S3Region != "" {
			o.Region = cfg.S3Region
		}
		if cfg.S3EndpointURL != "" {
			o.BaseEndpoint = &cfg.S3EndpointURL
		}
		o.UsePathStyle = cfg.S3UsePathStyle
	})
	s3FS := s3fs.New(s3Client, cfg.S3Bucket, s3fs.WithReadSeeker)
	return cacheClient.Middleware(http.FileServer(http.FS(s3FS)))
}
