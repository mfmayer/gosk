package llm

import "errors"

// Generator as a generic interface for large langage model generators
type Generator interface {
	// GenerateResponse to get response from the model behind the generator
	GenerateResponse(input Content) (response Content, err error)
}

// GeneratorConfig as a map of arbitrary generator configuration
// At least "model" should be available for each generator configuration
type GeneratorConfig map[string]interface{}

// Model returns the model name of the generator configuration. If not available "" is returned.
func (gc GeneratorConfig) Model() string {
	if gc == nil {
		return ""
	}
	if model, ok := gc["model"]; ok {
		if modelString, ok := model.(string); ok {
			return modelString
		}
	}
	return ""
}

// GeneratorMap as a map of generators
type GeneratorMap map[string]Generator

var (
	ErrUnknownGeneratorModel = errors.New("unknown generator model")
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
