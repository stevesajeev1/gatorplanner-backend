package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port string `env:"PORT" envDefault:"8080"`
}

func LoadConfig() *Config {
	loadEnv()

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatal().Msgf("Failed to parse env: %v", err)
	}

	return &cfg
}

func loadEnv() {
	files := []string{".env.local", ".env.dev", ".env"}
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			log.Info().Str("file", f).Msg("Loading environment file.")
			if err := godotenv.Load(f); err != nil {
				log.Fatal().Err(err).Msgf("Failed to load %s", f)
			}
			break
		}
	}
}
