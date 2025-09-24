package value

import (
	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type valueTransformer struct {
	Value any `yaml:"value"`
}

func init() {
	transformer.RegisterTransformerFactory("value", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &valueTransformer{}
		err := value.Decode(t)
		return t, err
	}))
}

func (this *valueTransformer) Transform(ctx *transformer.TransformationContext) error {
	ctx.Result = this.Value
	return nil
}
