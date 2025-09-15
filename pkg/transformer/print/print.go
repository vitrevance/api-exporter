package print

import (
	"log"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type printTransformer struct {
}

func init() {
	transformer.RegisterTransformerFactory("print", transformer.TransformerFactoryFunc(func(value *yaml.Node) (transformer.Transformer, error) {
		t := &printTransformer{}
		err := value.Decode(t)
		return t, err
	}))
}

func (this *printTransformer) Transform(ctx *transformer.TransformationContext) error {
	log.Println(ctx.Object)
	return nil
}
