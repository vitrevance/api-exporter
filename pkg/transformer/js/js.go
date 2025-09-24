package js

import (
	"fmt"

	"github.com/robertkrimen/otto"
	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

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
	value, err := vm.Run(this.Script)
	if err != nil {
		return err
	}
	ctx.Result, err = value.Export()
	return err
}
