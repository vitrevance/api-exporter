package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/vitrevance/api-exporter/pkg/fread"
	"github.com/vitrevance/api-exporter/pkg/runner"
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

func main() {
	configPath := flag.String("config", "config.yaml", "path to a config file")
	reloadIntervalStr := flag.String("reloadInterval", "0s", "config reload interval")
	flag.Parse()

	var reloadInterval time.Duration
	err := yaml.Unmarshal([]byte(*reloadIntervalStr), &reloadInterval)
	if err != nil {
		log.Fatalf("invalid reloadInterval format: %v", err)
	}

	cfgUpdates := reloadConfig(*configPath, reloadInterval)

	ctx, cancel := context.WithCancel(context.Background())
	for cfg := range cfgUpdates {
		cancel()
		ctx, cancel = context.WithCancel(context.Background())
		cfg.RunJobs(ctx)
	}
	cancel()
}

func reloadConfig(path string, reloadInterval time.Duration) <-chan *runner.Config {
	ch := make(chan *runner.Config)
	var lastConfig string
	var cfg *runner.Config
	go func() {
		defer close(ch)
		for {
			func() {
				bytes, err := fread.ReadFileOrHTTP(path)
				if err != nil {
					log.Printf("[ERROR] failed to reload config file: %v", err)
					time.Sleep(time.Second * 5)
					return
				}

				if lastConfig == string(bytes) {
					log.Println("[INFO] reloaded config with no changes")
					time.Sleep(reloadInterval)
					return
				}
				lastConfig = string(bytes)

				if cfg != nil {
					for k, _ := range cfg.Transformers {
						transformer.UnregisterTransformerFactory(k)
					}
				}

				cfg = &runner.Config{}
				err = yaml.Unmarshal(bytes, cfg)
				if err != nil {
					log.Printf("[ERROR] failed to read config: %v", err)
					time.Sleep(time.Second * 5)
					return
				}
				log.Println("[INFO] reloaded config")
				ch <- cfg
			}()
			if reloadInterval == 0 {
				return
			}
			time.Sleep(reloadInterval)
		}
	}()
	return ch
}
