package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BinanceWS string   `yaml:"BINANCE_WS"`
	Symbols   []string `yaml:"SYMBOLS"`

	Redis struct {
		Addr     string `yaml:"ADDR"`
		Password string `yaml:"PASSWORD"`
		DB       int    `yaml:"DB"`
	} `yaml:"REDIS"`

	TickTTLInt int `yaml:"TICK_TTL_SECONDS"`
	TickTTL    time.Duration

	CandleTTLInt int `yaml:"CANDLE_TTL_SECONDS"`
	CandleTTL    time.Duration
}

func (c *Config) ReadTTL() error {
	if c.TickTTLInt < 0 {
		return fmt.Errorf("tick TTL can not be negative, %v", c)
	}
	c.TickTTL = time.Duration(c.TickTTLInt) * time.Second

	if c.CandleTTLInt < 0 {
		return fmt.Errorf("candle TTL can not be negative, %v", c)
	}
	c.CandleTTL = time.Duration(c.CandleTTLInt) * time.Second

	return nil
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

	if err = config.ReadTTL(); err != nil {
		return nil, err
	}

	return &config, nil
}
