package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Mongo struct {
		Addr     string `yaml:"ADDR"`
		User string `yaml:"USER"`
		Password string `yaml:"PASS"`
	} `yaml:"MONGO"`

	SMTP struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		User      string `yaml:"user"`
		Pass      string `yaml:"pass"`
		FromEmail string `yaml:"from_email"`
	} `yaml:"SMTP"`

	JWTSecret string `yaml:"JWT_SECRET"`
}

func Load() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
