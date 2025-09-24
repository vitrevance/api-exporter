package sequence

import (
	"fmt"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type sequenceTransformer struct {
	Steps []transformer.TransformerConfig `yaml:"steps"`
}

func init() {
	transformer.RegisterTransformerFactory("sequence", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &sequenceTransformer{}
		err := value.Decode(t)
		return t, err
	}))
}

func (this *sequenceTransformer) Transform(ctx *transformer.TransformationContext) error {
	stepCtx := &transformer.TransformationContext{
		Object:       ctx.Object,
		Result:       ctx.Result,
		Transformers: ctx.Transformers,
	}
	for i, step := range this.Steps {
		if !step.KeepContext {
			if i > 0 {
				stepCtx = &transformer.TransformationContext{
					Object:       stepCtx.Result,
					Result:       make(map[string]any),
					Transformers: ctx.Transformers,
				}
			}
			if i+1 == len(this.Steps) {
				stepCtx.Result = ctx.Result
			}
		}
		err := step.Transformer.Transform(stepCtx)
		if err != nil {
			return fmt.Errorf("step [%d] failed: %v", i, err)
		}
	}
	ctx.Result = stepCtx.Result
	return nil
}
