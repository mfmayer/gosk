package llm

import (
	"errors"
	"fmt"
)

// GeneratorConfig to configure a specific generator's (defined by ID) response generator
type GeneratorConfig struct {
	TypeID           string                 `json:"typeID"`
	ConfigProperties map[string]interface{} `json:"config,omitempty"`
}

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

// GeneratorFactoryMap is a map of generator factories of different types
type GeneratorFactoryMap map[string]GeneratorFactory

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
	// ErrUnknownGeneratorModel is returned if a generator model is unknown
	ErrUnknownGeneratorModel = errors.New("unknown generator model")
	ErrUnknownGeneratorType  = errors.New("unknown generator type")
)

// FindAny returns first generator that matches any of the given patterns in given order
// func (m GeneratorMap) FindAny(patterns ...string) (Generator, error) {
// 	for _, pattern := range patterns {
// 		for k, g := range m {
// 			match, err := filepath.Match(pattern, k)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if match {
// 				return g, nil
// 			}
// 		}
// 	}
// 	return nil, fmt.Errorf("no generator found for patterns %v", patterns)
// }
