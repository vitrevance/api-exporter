package main

import (
	"flag"
	"log"
	"os"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"

	_ "github.com/vitrevance/api-exporter/pkg/transformer/array"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/field"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/http"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/js"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/parser"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/print"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/regex"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/sequence"
	_ "github.com/vitrevance/api-exporter/pkg/transformer/value"
)

type JobConfig struct {
	JobName string                          `yaml:"job_name"`
	Steps   []transformer.TransformerConfig `yaml:"steps"`
}

type Config struct {
	Transformers map[string]transformer.TransformerConfig `yaml:"transformers"`
	Jobs         []JobConfig                              `yaml:"jobs"`
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

	transformers := make(map[string]transformer.Transformer)
	for k, v := range cfg.Transformers {
		transformers[k] = v.Transformer
	}

	for _, job := range cfg.Jobs {
		log.Println("Starting job", job.JobName)
		ctx := &transformer.TransformationContext{
			Object:       make(map[string]any),
			Result:       make(map[string]any),
			Transformers: transformers,
		}
		for i, step := range job.Steps {
			if !step.KeepContext {
				ctx = &transformer.TransformationContext{
					Object:       ctx.Result,
					Result:       make(map[string]any),
					Transformers: transformers,
				}
			}
			err := step.Transformer.Transform(ctx)
			if err != nil {
				log.Printf("[ERROR] step [%d] failed: %v\n", i, err)
				break
			}
			log.Printf("[INFO] step [%d] finished\n", i)
		}
	}
}
