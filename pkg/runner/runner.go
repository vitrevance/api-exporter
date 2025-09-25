package runner

import (
	"context"
	"log"

	"github.com/vitrevance/api-exporter/pkg/transformer"
)

func (this *Config) RunJobs(ctx context.Context) {
	for _, job := range this.Jobs {
		go func() {
			for {
				log.Println("Starting job", job.JobName)
				tctx := &transformer.TransformationContext{
					Object:       make(map[string]any),
					Result:       make(map[string]any),
					Transformers: this.Transformers,
				}
				for i, step := range job.Steps {
					if !step.KeepContext {
						tctx = &transformer.TransformationContext{
							Object:       tctx.Result,
							Result:       make(map[string]any),
							Transformers: this.Transformers,
						}
					}
					err := step.Transformer.Transform(tctx)
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
				timeout, cancel := context.WithTimeout(context.Background(), job.RunInterval)
				select {
				case <-ctx.Done():
					cancel()
					return
				case <-timeout.Done():
					cancel()
				}
			}
		}()
	}
}
