package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	err := godotenv.Load() //tries to read a file named .env

	if err != nil {
		log.Println("Warning: .env file not found, using environmental variables")
	}

	config := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}
	return config, nil
}
