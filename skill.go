package gosk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/mfmayer/gosk/pkg/llm"
	"golang.org/x/exp/slices"
)

// Skill defines and holds a collection of Skill Functions that can be planned and called by the semantic kernel
type Skill struct {
	// Name of the skill
	Name string `json:"name,omitempty"`
	// Description of the skill should indicate and give an idea about the functions that are included
	Description string `json:"description"`
	// Functions that the skill provides
	Functions map[string]*Function `json:"functions"`
}

// Call a skill function with given name and parameters
func (s *Skill) Call(functionName string, input llm.Content) (response llm.Content, err error) {
	function, ok := s.Functions[functionName]
	if !ok {
		return nil, fmt.Errorf("function `%s` not found", functionName)
	}
	// Check input for required parameters and eventually set default values
	for _, parameter := range function.Parameters {
		if parameter.Default != nil {
			if _, ok := input.Option(parameter.Name); !ok {
				input.With(parameter.Name, parameter.Default)
			}
		}
		if parameter.Required {
			if _, ok := input.Option(parameter.Name); !ok {
				err = errors.Join(err, fmt.Errorf("parameter `%s` is required", parameter.Name))
			}
		}

	}
	if err != nil {
		return nil, err
	}
	return function.Call(input)
}

// ParseSemanticSkillFromFS parses a skill from a file system
// fsys is the file system to parse the skill from (see assets/skills for examples)
// generators is a map of generators that can be used by the skill
func ParseSemanticSkillFromFS(fsys fs.FS, generators map[string]llm.Generator) (skill *Skill, err error) {
	// open config file
	file, err := fsys.Open("config.json")
	if err != nil {
		err = fmt.Errorf("opening `config.json` failed: %w", err)
		return
	}
	defer file.Close()

	// read config file
	data, err := ioutil.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("reading `config.json` failed: %w", err)
		return
	}

	// unmarshal config file
	var s Skill
	err = json.Unmarshal(data, &s)
	if err != nil {
		err = fmt.Errorf("unmarshalling `config.json` failed: %w", err)
		return
	}
	skill = &s
	skill.Functions = map[string]*Function{}

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
		// // skip root directory
		// if d.path == "." {
		// 	continue
		// }
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

// ParseSemanticSkillsFromFS parses all skills from a file system
// File system should be a directory with subdirectories for each skill
func ParseSemanticSkillsFromFS(fsys fs.FS, generators map[string]llm.Generator, dirs ...string) (skills map[string]*Skill, err error) {
	// create slice for skills
	skills = map[string]*Skill{}
	// fsys := os.DirFS(path)
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		err = fmt.Errorf("reading file system failed: %w", err)
		return
	}

	for _, d := range entries {
		// skip files
		if !d.IsDir() {
			continue
		}
		if len(dirs) > 0 {
			// if dirs are given, skip directories that are not in dirs
			if !slices.Contains(dirs, d.Name()) {
				continue
			}
		}
		// create subFS for subdirectory
		subFS, err := fs.Sub(fsys, d.Name())
		if err != nil {
			// skip subdirectory if it is not possible to create subFS
			continue
		}
		// create skill from subFS
		skill, err := ParseSemanticSkillFromFS(subFS, generators)
		if err != nil {
			// skip subdirectory if it is not possible to create skill
			continue
		}
		// add skill to skills with its directory name as key
		skills[strings.ToLower(d.Name())] = skill
	}
	return
}
