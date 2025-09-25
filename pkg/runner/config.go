package runner

import (
	"time"

	"github.com/vitrevance/api-exporter/pkg/transformer"
	"gopkg.in/yaml.v3"
)

type JobConfig struct {
	JobName     string                          `yaml:"job_name"`
	RunInterval time.Duration                   `yaml:"interval"`
	Steps       []transformer.TransformerConfig `yaml:"steps"`
}

type Config struct {
	Transformers map[string]transformer.Transformer `yaml:"-"`
	Jobs         []JobConfig                        `yaml:"-"`
}

func (this *Config) UnmarshalYAML(value *yaml.Node) error {
	this.Transformers = make(map[string]transformer.Transformer)

	type helperT struct {
		Transformers map[string]*yaml.Node `yaml:"transformers"`
	}
	transformers := helperT{}
	err := value.Decode(&transformers)
	if err != nil {
		return err
	}
	for k, _ := range transformers.Transformers {
		err = transformer.RegisterTransformerFactory(k, transformer.NewAliasTransformerFactory(k))
		if err != nil {
			return err
		}
	}

	type helper struct {
		Transformers map[string]transformer.TransformerConfig `yaml:"transformers"`
		Jobs         []JobConfig                              `yaml:"jobs"`
	}
	h := helper{}
	err = value.Decode(&h)
	if err != nil {
		return err
	}
	for k, v := range h.Transformers {
		this.Transformers[k] = v.Transformer
	}
	this.Jobs = h.Jobs
	return nil
}
