package transformer

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type TransformationContext struct {
	Object       any
	Result       any
	Transformers map[string]Transformer
}

type Transformer interface {
	Transform(*TransformationContext) error
}

type TransformerFactory interface {
	UnmarshalYAML(value *yaml.Node) (Transformer, error)
}

type TransformerFactoryFunc func(value *yaml.Node) (Transformer, error)

func (this TransformerFactoryFunc) UnmarshalYAML(value *yaml.Node) (Transformer, error) {
	return this(value)
}

var transformerTypes map[string]TransformerFactory = make(map[string]TransformerFactory)

func RegisterTransformerFactory(name string, factory TransformerFactory) error {
	if transformerTypes[name] != nil {
		return fmt.Errorf("transformer with name %s is already registered", name)
	}
	transformerTypes[name] = factory
	return nil
}

type TransformerConfig struct {
	Type        string
	KeepContext bool
	Transformer Transformer
}

func (this *TransformerConfig) UnmarshalYAML(value *yaml.Node) error {
	type typeHeader struct {
		Type        string `yaml:"type"`
		KeepContext bool   `yaml:"keep_ctx"`
	}
	t := &typeHeader{}
	err := value.Decode(t)
	if err != nil {
		return err
	}
	if t.Type == "" {
		return fmt.Errorf("transformer type is required")
	}
	this.Type = t.Type
	this.KeepContext = t.KeepContext
	factory := transformerTypes[this.Type]

	if factory == nil {
		return fmt.Errorf("unknown transformer type %s", this.Type)
	}

	tr, err := factory.UnmarshalYAML(value)
	if err != nil {
		return err
	}
	this.Transformer = tr
	return nil
}
