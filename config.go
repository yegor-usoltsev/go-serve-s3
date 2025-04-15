package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

const envPrefix = "APP"

type Config struct {
	ServerHost           string        `split_words:"true" required:"true" default:"0.0.0.0"`
	ServerPort           uint16        `split_words:"true" required:"true" default:"8080"`
	S3Bucket             string        `split_words:"true" required:"true"`
	S3Region             string        `split_words:"true" required:"false"`
	S3EndpointURL        string        `split_words:"true" required:"false"`
	S3UsePathStyle       bool          `split_words:"true" required:"false"`
	CachingCapacityItems int           `split_words:"true" required:"true" default:"1024"`
	CachingCapacityBytes int           `split_words:"true" required:"true" default:"52428800"` // 50 MiB
	CachingTTL           time.Duration `split_words:"true" required:"true" default:"10m"`      // 10 minutes
}

func NewConfigFromEnv() Config {
	var cfg Config
	if err := envconfig.Process(envPrefix, &cfg); err != nil {
		_ = envconfig.Usage(envPrefix, &cfg)
		panic(err)
	}
	return cfg
}
