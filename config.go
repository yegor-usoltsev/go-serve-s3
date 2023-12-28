package main

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

const envPrefix = "app"

type Config struct {
	ServerHost string `split_words:"true" default:"localhost"`
	ServerPort uint16 `split_words:"true" default:"8080"`
	S3Bucket   string `split_words:"true" required:"true"`
	S3Region   string `split_words:"true" required:"true"`
}

func NewConfig() Config {
	var cfg Config
	if err := envconfig.Process(envPrefix, &cfg); err != nil {
		_ = envconfig.Usage(envPrefix, &cfg)
		fmt.Printf("\n%v\n", err)
		os.Exit(1)
	}
	return cfg
}
