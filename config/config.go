package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

type BinanceConfig struct {
	APIURL       string `yaml:"api_url"`
	TimeoutSec   int    `yaml:"timeout_seconds"`
	Convertation string `yaml:"convertation"`
}

// Load загружает конфиг из YAML файла
func Load(configPath string) (*BinanceConfig, error) {
	fullPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении пути до конфига: %w", err)
	}

	yamlFile, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("ошибки при чтении конфига: %w", err)
	}

	var cfg BinanceConfig
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка при переводе конфига: %w", err)
	}

	log.Printf("Домен для запросов на цены криптовалют: %s", cfg.APIURL)
	log.Printf("Опорная валюта для конвертации валют: '%v'", cfg.Convertation)

	return &cfg, nil
}
