package llm

import (
	"encoding/json"
	"errors"
	"fmt"
)

// RegistrationFunc is used to register a new type of generator with the go semantic kernel (gosk)
type RegistrationFunc func() (typeID string, newGenerator NewGeneratorFunc)

// NewGeneratorFunc creates a new generator with given config
type NewGeneratorFunc func(config GeneratorConfigData) (Generator, error)

// // GeneratorFactory allows to create response generators
// type GeneratorFactory interface {
// 	TypeID() string
// 	New(config GeneratorConfigData) (Generator, error)
// }

// Generator as a generic interface for large langage model response generators
type Generator interface {
	// GenerateResponse to get response from the model behind the generator
	Generate(input Content) (response Content, err error)
}

// GeneratorConfig to configure a specific generator's (defined by ID) response generator
type GeneratorConfig struct {
	TypeID           string              `json:"typeID"`
	ConfigProperties GeneratorConfigData `json:"config,omitempty"`
}

type GeneratorConfigData map[string]interface{}

// Convert config data into given object
func (gcd *GeneratorConfigData) Convert(to interface{}) (err error) {
	if gcd == nil {
		return
	}
	jsonData, err := json.Marshal(gcd)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, to)
	if err != nil {
		return err
	}
	return
}

// NewGeneratorFuncMap is a map of generator factories of different types
type NewGeneratorFuncMap map[string]NewGeneratorFunc

func (gm NewGeneratorFuncMap) CreateGenerator(typeID string, config map[string]interface{}) (Generator, error) {
	newGeneratorFunc, ok := gm[typeID]
	if !ok {
		return nil, fmt.Errorf("%w: `%s`", ErrUnknownGeneratorType, typeID)
	}
	return newGeneratorFunc(config)
}

// CreateResponseGenerators creates response generators map from a given config map. Their keys are the names of the response generators.
func (gm NewGeneratorFuncMap) CreateGenerators(generatorConfigs map[string]GeneratorConfig) (generators map[string]Generator, err error) {
	generators = map[string]Generator{}
	for generatorName, generatorConfig := range generatorConfigs {
		newGeneratorFunc, ok := gm[generatorConfig.TypeID]
		if !ok {
			err = errors.Join(err, fmt.Errorf("%w: `%s`", ErrUnknownGeneratorType, generatorConfig.TypeID))
			continue
		}
		generator, newGenError := newGeneratorFunc(generatorConfig.ConfigProperties)
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
