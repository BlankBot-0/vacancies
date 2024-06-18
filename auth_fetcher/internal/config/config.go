package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	LoginURL    string      `yaml:"login_url"`
	Credentials Credentials `yaml:"credentials"`
	Captcha     Captcha     `yaml:"captcha"`
}

type Credentials struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type Captcha struct {
	CaptchaInURL     string `yaml:"captcha_in_url"`
	CaptchaResultURL string `yaml:"captcha_result_url"`
	ApiKey           string `yaml:"api_key"`
	SiteKey          string `yaml:"site_key"`
	PageUrl          string `yaml:"page_url"`
	JsonFlag         string `yaml:"json_flag"`
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
