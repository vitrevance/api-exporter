package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

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
	JobName     string                          `yaml:"job_name"`
	RunInterval time.Duration                   `yaml:"interval"`
	Steps       []transformer.TransformerConfig `yaml:"steps"`
}

type Config struct {
	Transformers map[string]transformer.TransformerConfig `yaml:"transformers"`
	Jobs         []JobConfig                              `yaml:"jobs"`
}

func (this *Config) UnmarshalYAML(value *yaml.Node) error {
	this.Transformers = make(map[string]transformer.TransformerConfig)

	type helperT struct {
		Transformers map[string]*yaml.Node `yaml:"transformers"`
	}
	transformers := helperT{}
	err := value.Decode(&transformers)
	if err != nil {
		return err
	}
	for k, _ := range transformers.Transformers {
		err = transformer.RegisterTransformerFactory(k, transformer.NewAliasTransformerFactory(k))
		if err != nil {
			return err
		}
	}

	type helper struct {
		Transformers map[string]transformer.TransformerConfig `yaml:"transformers"`
		Jobs         []JobConfig                              `yaml:"jobs"`
	}
	h := helper{}
	err = value.Decode(&h)
	if err != nil {
		return err
	}
	this.Transformers = h.Transformers
	this.Jobs = h.Jobs
	return nil
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

	wg := sync.WaitGroup{}

	for _, job := range cfg.Jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
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
				log.Println("Finished job", job.JobName)
				if job.RunInterval == 0 {
					return
				}
				time.Sleep(job.RunInterval)
			}
		}()
	}

	wg.Wait()
}
