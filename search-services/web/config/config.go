package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"WEB_ADDRESS" env-default:"localhost:8080"`
	Timeout time.Duration `yaml:"timeout" env:"WEB_TIMEOUT" env-default:"10m"`
}

type Config struct {
	LogLevel   string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTPConfig HTTPConfig `yaml:"web_server"`
	APIAddress string     `yaml:"api_address" env:"API_ADDRESS" env-default:"http://localhost:28080"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
