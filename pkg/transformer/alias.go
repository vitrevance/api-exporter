package transformer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type aliasTransformer struct {
	Ctx         map[string]any `yaml:"ctx"`
	transformer string
}

func NewAliasTransformerFactory(transformer string) TransformerFactory {
	return TransformerFactoryFunc(func(value *yaml.Node) (Transformer, error) {
		t := &aliasTransformer{}
		err := value.Decode(t)
		if err != nil {
			return nil, err
		}
		t.transformer = transformer
		return t, nil
	})
}

func (this *aliasTransformer) Transform(ctx *TransformationContext) error {
	if len(this.Ctx) > 0 {
		obj, ok := ctx.Object.(map[string]any)
		if !ok {
			return fmt.Errorf("cannot merge non-Object context")
		}
		for k, v := range this.Ctx {
			obj[k] = v
		}
	}
	tr, ok := ctx.Transformers[this.transformer]
	if !ok {
		return fmt.Errorf("alias for undefined transformer: %s", this.transformer)
	}
	return tr.Transform(ctx)
}
