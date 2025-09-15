package main

import (
	"flag"
	"log"
	"os"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"

	_ "github.com/vitrevance/api-exporter/pkg/transformer/field"
)

type Config struct {
	Transformers map[string]transformer.TransformerConfig `yaml:"transformers"`
}

func main() {
	configPath := flag.String("config", "config.yaml", "path to a config file")
	flag.Parse()

	bytes, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(bytes, cfg)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
}
