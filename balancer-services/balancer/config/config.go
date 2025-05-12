package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

type HelloConfig struct {
	FirstAddress  string `yaml:"first_address" env:"FIRST_ADDRESS"`
	SecondAddress string `yaml:"second_address" env:"SECOND_ADDRESS"`
	ThirdAddress  string `yaml:"third_address" env:"THIRD_ADDRESS"`
}
type HTTPConfig struct {
	Address string        `yaml:"address" env:"BALANCER_ADDRESS" env-default:"localhost:80"`
	Timeout time.Duration `yaml:"timeout" env:"BALANCER_TIMEOUT" env-default:"5s"`
}

type Config struct {
	LogLevel    string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Concurrency int           `yaml:"concurrency" env:"CONCURRENCY" env-default:"10"`
	RateLimit   int           `yaml:"rate_limit" env:"RATE_LIMIT" env-default:"3"`
	RateTime    time.Duration `yaml:"rate_time" env:"RATE_TIME" env-default:"30s"`
	HelloConfig HelloConfig   `yaml:"hello"`
	HTTPConfig  HTTPConfig    `yaml:"balancer"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
