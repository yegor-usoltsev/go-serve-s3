package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jszwec/s3fs/v2"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

const (
	cachingAlgorithm     = memory.LRU
	cachingCapacityItems = 1024
	cachingCapacityBytes = 50 * 1024 * 1024 // 50 MiB
	cachingTTL           = 10 * time.Minute
)

type Handler struct {
	*http.ServeMux
}

func NewHandler(cfg Config) *Handler {
	h := &Handler{ServeMux: http.NewServeMux()}
	h.HandleFunc("/health", healthHandler)
	h.Handle("/", s3Handler(cfg.S3Bucket, cfg.S3Region))
	return h
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprint(w, "OK")
}

func s3Handler(bucket, region string) http.Handler {
	memoryAdapter, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(cachingAlgorithm),
		memory.AdapterWithCapacity(cachingCapacityItems),
		memory.AdapterWithStorageCapacity(cachingCapacityBytes),
	)
	if err != nil {
		log.Fatal(err)
	}
	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memoryAdapter),
		cache.ClientWithExpiresHeader(),
		cache.ClientWithTTL(cachingTTL),
	)
	if err != nil {
		log.Fatal(err)
	}
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Print("Couldn't load AWS configuration. Have you set up your AWS account?")
		log.Fatal(err)
	}
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) { o.Region = region })
	s3FS := s3fs.New(s3Client, bucket)
	s3fs.WithReadSeeker(s3FS)
	return cacheClient.Middleware(http.FileServer(http.FS(s3FS)))
}
