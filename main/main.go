package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	addr := flag.String("addr", "", "address to listen on")
	flag.Parse()

	var reloadInterval time.Duration
	err := yaml.Unmarshal([]byte(*reloadIntervalStr), &reloadInterval)
	if err != nil {
		log.Fatalf("invalid reloadInterval format: %v", err)
	}

	cfgUpdates := reloadConfig(*configPath, reloadInterval)

	if *addr != "" {
		// Start HTTP server
		var cfgHandler *runner.Config
		mux := sync.Mutex{}
		go func() {
			log.Printf("Starting HTTP server on %s", *addr)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mux.Lock()
				cfg := cfgHandler
				mux.Unlock()
				if cfg != nil {
					cfg.Handle(w, r)
				}
			})
			if err := http.ListenAndServe(*addr, handler); err != nil {
				log.Fatalf("Failed to start HTTP server: %v", err)
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			for cfg := range cfgUpdates {
				cancel()
				mux.Lock()
				cfgHandler = cfg
				mux.Unlock()
				ctx, cancel = context.WithCancel(context.Background())
				cfg.RunJobs(ctx)
			}
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		log.Println("Server started, press Ctrl+C to stop")
		<-sigChan

		log.Println("Shutting down...")
		cancel()
	} else {
		// Original behavior when no address is specified
		ctx, cancel := context.WithCancel(context.Background())
		for cfg := range cfgUpdates {
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			cfg.RunJobs(ctx)
		}
		cancel()
	}
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
