package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	// Google credentials
	EmailID       string `env:"EMAIL_ID,required"`
	EmailPassword string `env:"EMAIL_PASSWORD,required"`

	// Meeting
	MeetLink          string `env:"MEET_LINK"`
	RecordingDuration int    `env:"RECORDING_DURATION" envDefault:"60"`

	// Audio
	SampleRate        int `env:"SAMPLE_RATE" envDefault:"44100"`
	MaxAudioSizeBytes int `env:"MAX_AUDIO_SIZE_BYTES" envDefault:"20971520"`

	// OpenAI
	OpenAIAPIKey string `env:"OPENAI_API_KEY,required"`
	GPTModel     string `env:"GPT_MODEL" envDefault:"gpt-4"`
	WhisperModel string `env:"WHISPER_MODEL" envDefault:"whisper-1"`

	// Server
	Port string `env:"PORT" envDefault:"8080"`
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if missing)
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
