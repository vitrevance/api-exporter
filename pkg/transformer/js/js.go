package js

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/moby/moby/client"
	"github.com/robertkrimen/otto"
	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type ExecutionContext struct {
	DockerClient *client.Client
}

var GlobalExecutionContext = ExecutionContext{}

type jsTransformer struct {
	Script string `yaml:"script"`
}

func init() {
	transformer.RegisterTransformerFactory("javascript", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &jsTransformer{}
		err := value.Decode(t)
		t.Script = fmt.Sprintf("(function(){\n%s\n})()", t.Script)
		return t, err
	}))
}

func (this *jsTransformer) Transform(ctx *transformer.TransformationContext) error {
	vm := otto.New()
	vm.Set("source", ctx.Object)
	vm.Set("target", ctx.Result)
	vm.Set("run", func(name string, args any) any {
		tr := ctx.Transformers[name]
		if tr != nil {
			taskCtx := &transformer.TransformationContext{
				Object:       args,
				Result:       make(map[string]any),
				Transformers: ctx.Transformers,
			}
			err := tr.Transform(taskCtx)
			if err != nil {
				return map[string]any{"error": err.Error()}
			}
			return taskCtx.Result
		}
		return map[string]any{"error": "undefined transformer"}
	})
	vm.Set("docker_container_list", func(filters map[string][]string, all bool, limit int) any {
		if GlobalExecutionContext.DockerClient == nil {
			return map[string]any{"error": "docker is not enabled"}
		}
		opts := client.ContainerListOptions{
			All:     all,
			Limit:   limit,
			Filters: make(client.Filters),
		}
		for k, v := range filters {
			opts.Filters.Add(k, v...)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		containers, err := GlobalExecutionContext.DockerClient.ContainerList(ctx, opts)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return containers.Items
	})
	vm.Set("docker_container_inspect", func(id string) any {
		if GlobalExecutionContext.DockerClient == nil {
			return map[string]any{"error": "docker is not enabled"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		container, err := GlobalExecutionContext.DockerClient.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		m := make(map[string]any)
		err = json.Unmarshal(container.Raw, &m)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return m
	})
	vm.Set("docker_container_kill", func(id string, sig string) any {
		if GlobalExecutionContext.DockerClient == nil {
			return map[string]any{"error": "docker is not enabled"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := GlobalExecutionContext.DockerClient.ContainerKill(ctx, id, client.ContainerKillOptions{
			Signal: sig,
		})
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return nil
	})
	vm.Set("docker_container_restart", func(id string, sig string) any {
		if GlobalExecutionContext.DockerClient == nil {
			return map[string]any{"error": "docker is not enabled"}
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, err := GlobalExecutionContext.DockerClient.ContainerRestart(ctx, id, client.ContainerRestartOptions{
			Signal: sig,
		})
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return nil
	})
	vm.Set("getenv", func(id string) any {
		return os.Getenv(id)
	})
	value, err := vm.Run(this.Script)
	if err != nil {
		return err
	}
	ctx.Result, err = value.Export()
	return err
}
