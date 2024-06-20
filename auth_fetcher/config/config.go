package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Credentials   Credentials `yaml:"credentials"`
	CaptchaApiKey string      `yaml:"captcha_api_key"`
	Port          string      `yaml:"port"`
}

type Credentials struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

func MustLoad() *Config {
	envFile, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	configPath := envFile["AUTH_CONFIG_PATH"]
	if configPath == "" {
		log.Fatal("AUTH_CONFIG_PATH environment variable not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}
	return &cfg
}
