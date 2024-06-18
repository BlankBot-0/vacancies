package internal

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	VacanciesURL string `yaml:"vacancies_url"`
	KeyWord      string `yaml:"key_word"`
}

func MustLoad() *Config {
	envFile, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	configPath := envFile["FETCHER_CONFIG_PATH"]
	if configPath == "" {
		log.Fatal("FETCHER_CONFIG_PATH environment variable not set")
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
