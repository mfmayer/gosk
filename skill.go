package gosk

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"github.com/mfmayer/gosk/pkg/llm"
)

// SkillFactoryFunc to create a semantic skill with the help of given generators
type SkillFactoryFunc func(generatorFactories llm.GeneratorFactoryMap) (skill *Skill, err error)

// Skill defines and holds a collection of Skill Functions that can be planned and called by the semantic kernel
type Skill struct {
	// Name of the skill
	Name string `json:"name,omitempty"`
	// Description of the skill should indicate and give an idea about the functions that are included
	Description string `json:"description"`
	// Plannable indicates whether the skill can be planned by the semantic kernel
	Plannable bool `json:"plannable,omitempty"`
	// Functions that the skill provides
	Functions map[string]*Function `json:"functions"`
	// Generators that might be initialized while parsing a skill from a configuration
	Generators map[string]llm.Generator `json:"-"`
}

func (s *Skill) String() string {
	if s == nil {
		return "undefined skill"
	}
	return fmt.Sprintf("skill `%s`", s.Name)
}

// Call a skill function with given name and input content
// func (s *Skill) Call(functionName string, input llm.Content) (response llm.Content, err error) {
// 	function, ok := s.Functions[functionName]
// 	if !ok {
// 		return nil, fmt.Errorf("function `%s` not found", functionName)
// 	}
// 	// Check input for required input properties and eventually set default values
// 	for _, parameter := range function.InputProperties {
// 		if parameter.Default != nil {
// 			if input.Property(parameter.Name) == nil {
// 				input.With(parameter.Name, parameter.Default)
// 			}
// 		}
// 		if parameter.Required {
// 			if input.Property(parameter.Name) == nil {
// 				err = errors.Join(err, fmt.Errorf("%w: `%s`", ErrMissingParameter, parameter.Name))
// 			}
// 		}

// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return function.Call(input)
// }

type skillConfig struct {
	*Skill
	GeneratorConfigs map[string]llm.GeneratorConfig `json:"generators,omitempty"`
}

// ParseSemanticSkillFromFS parses a skill from fsys file system (see assets/skills for examples).
// Given generatorFactories are used to create and return generators that are configured for this skill.
func ParseSemanticSkillFromFS(fsys fs.FS, generatorFactories llm.GeneratorFactoryMap, options ...createSemanticFunctionsOption) (skill *Skill, err error) {
	// open config file
	file, err := fsys.Open("config.json")
	if err != nil {
		err = fmt.Errorf("opening `config.json` failed: %w", err)
		return
	}
	defer file.Close()

	// read config file
	data, err := io.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("reading `config.json` failed: %w", err)
		return
	}

	// unmarshal config file
	var skillConfig skillConfig
	err = json.Unmarshal(data, &skillConfig)
	if err != nil {
		err = fmt.Errorf("unmarshalling `config.json` failed: %w", err)
		return
	}
	skill = skillConfig.Skill
	if skill == nil {
		err = fmt.Errorf("invalid skill `config.json`")
		return
	}

	// create response generators
	generators, err := generatorFactories.CreateGenerators(skillConfig.GeneratorConfigs)
	if err != nil {
		err = fmt.Errorf("creating generators failed: %w", err)
		return
	}

	// create configured skill functions
	functions, err := ParseSemanticFunctionsFromFS(fsys, generators, options...)
	if err != nil {
		err = fmt.Errorf("parsing functions failed: %w", err)
		return
	}
	skill.Functions = functions
	skill.Generators = generators

	return
}
