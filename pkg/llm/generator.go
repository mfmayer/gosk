package llm

import (
	"errors"
	"fmt"
)

// GeneratorFactory allows to create response generators
type GeneratorFactory interface {
	TypeID() string
	New(config map[string]interface{}) (Generator, error)
}

// Generator as a generic interface for large langage model response generators
type Generator interface {
	// GenerateResponse to get response from the model behind the generator
	Generate(input Content) (response Content, err error)
}

// GeneratorConfig to configure a specific generator's (defined by ID) response generator
type GeneratorConfig struct {
	TypeID           string                 `json:"typeID"`
	ConfigProperties map[string]interface{} `json:"config,omitempty"`
}

// GeneratorFactoryMap is a map of generator factories of different types
type GeneratorFactoryMap map[string]GeneratorFactory

func (gm GeneratorFactoryMap) CreateGenerator(typeID string, config map[string]interface{}) (Generator, error) {
	generatorFactory, ok := gm[typeID]
	if !ok {
		return nil, fmt.Errorf("%w: `%s`", ErrUnknownGeneratorType, typeID)
	}
	return generatorFactory.New(config)
}

// CreateResponseGenerators creates response generators map from a given config map. Their keys are the names of the response generators.
func (gm GeneratorFactoryMap) CreateGenerators(generatorConfigs map[string]GeneratorConfig) (generators map[string]Generator, err error) {
	generators = map[string]Generator{}
	for generatorName, generatorConfig := range generatorConfigs {
		generatorFactory, ok := gm[generatorConfig.TypeID]
		if !ok {
			err = errors.Join(err, fmt.Errorf("%w: `%s`", ErrUnknownGeneratorType, generatorConfig.TypeID))
			continue
		}
		generator, newGenError := generatorFactory.New(generatorConfig.ConfigProperties)
		if newGenError != nil {
			err = errors.Join(err, fmt.Errorf("creating generator \"%s\" failed: %w", generatorName, newGenError))
			continue
		}
		generators[generatorName] = generator
	}
	return
}

var (
	ErrUnknownGeneratorType = errors.New("unknown generator type")
)
