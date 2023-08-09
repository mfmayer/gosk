package gosk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/mfmayer/gosk/pkg/llm"
)

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
}

// Call a skill function with given name and input content
func (s *Skill) Call(functionName string, input llm.Content) (response llm.Content, err error) {
	function, ok := s.Functions[functionName]
	if !ok {
		return nil, fmt.Errorf("function `%s` not found", functionName)
	}
	// Check input for required input properties and eventually set default values
	for _, parameter := range function.InputProperties {
		if parameter.Default != nil {
			if input.Property(parameter.Name) == nil {
				input.With(parameter.Name, parameter.Default)
			}
		}
		if parameter.Required {
			if input.Property(parameter.Name) == nil {
				err = errors.Join(err, fmt.Errorf("%w: `%s`", ErrMissingParameter, parameter.Name))
			}
		}

	}
	if err != nil {
		return nil, err
	}
	return function.Call(input)
}

type skillConfig struct {
	*Skill
	Generators map[string]llm.GeneratorConfig `json:"generators"`
}

// ParseSemanticSkillFromFS parses a skill from a file system
// fsys is the file system to parse the skill from (see assets/skills for examples)
// generators is a map of generators that can be used by the skill
func ParseSemanticSkillFromFS(fsys fs.FS, getGenerators func(generatorConfigs map[string]llm.GeneratorConfig) (llm.GeneratorMap, error)) (skill *Skill, err error) {
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
	skill.Functions = map[string]*Function{}

	// get configured generators
	generators, err := getGenerators(skillConfig.Generators)
	if err != nil {
		err = fmt.Errorf("getting generators failed: %w", err)
		return
	}

	// find and parse skill functions in sub directories
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		err = fmt.Errorf("reading file system failed: %w", err)
		return
	}
	for _, d := range entries {
		if !d.IsDir() {
			continue
		}
		// create subFS for subdirectory
		subFS, err := fs.Sub(fsys, d.Name())
		if err != nil {
			// skip subdirectory if it is not possible to create subFS
			continue
		}
		// create function from subFS
		function, err := ParseSemanticFunctionFromFS(subFS, generators)
		if err != nil {
			// skip subdirectory if it is not possible to create function
			continue
		}
		// add function to skill with its directory name as key
		skill.Functions[strings.ToLower(d.Name())] = function
	}

	return
}
