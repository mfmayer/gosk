package gosk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mfmayer/gosk/pkg/llm"
)

var (
	// ErrMissingParameter is returned when a required parameter is missing
	ErrMissingParameter = errors.New("missing parameter")
)

// SemanticKernel
type SemanticKernel struct {
	// generators map[string]llm.Generator
	skills map[string]*Skill
}

type newKernelOption func(*newKernelOptions)

type newKernelOptions struct {
}

// WithOpenAIKey to use this OpenAI key when creating a new semantic kernel, otherwise it's tried to get the key from "OPENAI_API_KEY" environment variable or .env file in current working directory
// func WithOpenAIKey(key string) newKernelOption {
// 	return func(opt *newKernelOptions) {
// 		opt.openAIKey = key
// 	}
// }

// NewKernel creates new kernel and tries to retrieve the OpenAI key from "OPENAI_API_KEY" environment variable or .env file in current working directory
func NewKernel(opts ...newKernelOption) *SemanticKernel {
	options := &newKernelOptions{}
	for _, opt := range opts {
		opt(options)
	}

	kernel := &SemanticKernel{
		skills: map[string]*Skill{},
	}
	return kernel
}

// AddSkillsMap adds skills map to the kernel with adopted keys as skill names
// func (sk *SemanticKernel) AddSkillsMap(skills map[string]*Skill) (err error) {
// 	for skillName, skill := range skills {
// 		err = errors.Join(err, sk.addSkill(skillName, skill))
// 	}
// 	return
// }

// CreateAndAddSkills creates new skills and adds them to the kernel with their individual names
func (sk *SemanticKernel) CreateAndAddSkills(newSkillFunctions ...func() (skill *Skill, err error)) (err error) {
	for _, newSkillFunc := range newSkillFunctions {
		skill, newSkillErr := newSkillFunc()
		if newSkillErr != nil {
			err = errors.Join(err, newSkillErr)
			continue
		}
		err = errors.Join(err, sk.addSkill(skill.Name, skill))
	}
	return
}

// AddSkills adds skills to the kernel with their individual names
func (sk *SemanticKernel) AddSkills(skills ...*Skill) (err error) {
	for _, skill := range skills {
		err = errors.Join(err, sk.addSkill(skill.Name, skill))
	}
	return nil
}

func (sk *SemanticKernel) addSkill(name string, skill *Skill) error {
	// sanitize skill and parameter names
	if skill == nil {
		return fmt.Errorf("skill `%s` is nil", name)
	}
	if skill.Name == "" {
		skill.Name = name
	}
	for functionName, function := range skill.Functions {
		if function.Name == "" {
			function.Name = functionName
		}
		for parameterName, parameter := range function.Parameters {
			if parameter.Name == "" {
				parameter.Name = parameterName
			}
		}
	}
	// check if skill already exists
	if _, ok := sk.skills[name]; ok {
		return fmt.Errorf("skill `%s` already added", name)
	}
	sk.skills[name] = skill
	return nil
}

// FindSkill finds a skill by name and returns it or an error if not found
func (sk *SemanticKernel) FindSkill(skillName string) (skill *Skill, ok bool) {
	if skill, ok := sk.skills[skillName]; ok {
		return skill, true
	}
	return nil, false
}

// FindFunction finds a function in a skill by name and returns it or an error if not found
func (sk *SemanticKernel) FindFunction(skillName string, skillFunction string) (function *Function, ok bool) {
	if skill, ok := sk.FindSkill(skillName); ok {
		if function, ok := skill.Functions[skillFunction]; ok {
			return function, true
		}
	}
	return nil, false
}

// FindFunctions finds functions with path notation (`skillName.functionName`) and returns them or an error if any function is not found
func (sk *SemanticKernel) FindFunctions(functionPaths ...string) (functions []*Function, err error) {
	if len(functionPaths) == 0 {
		return nil, fmt.Errorf("no function path given")
	}
	functions = make([]*Function, 0, len(functionPaths))
	pathsNotFound := []string{}
	for _, fp := range functionPaths {
		fps := strings.Split(fp, ".")
		if len(fps) != 2 {
			pathsNotFound = append(pathsNotFound, fp)
			continue
		}
		if function, ok := sk.FindFunction(fps[0], fps[1]); ok {
			functions = append(functions, function)
		} else {
			pathsNotFound = append(pathsNotFound, fp)
		}
	}
	if len(pathsNotFound) > 0 {
		err = fmt.Errorf("functions %v not found", pathsNotFound)
	}
	return
}

// Call a skill function with given name and parameters and returns the response and/or an error
// The kernel also links the input as predecessor to the response
func (sk *SemanticKernel) Call(skillName string, skillFunction string, input llm.Content) (response llm.Content, err error) {
	if skill, ok := sk.skills[skillName]; ok {
		response, err = skill.Call(skillFunction, input)
		if response != nil {
			response.WithPredecessor(input)
		}
		return
	}
	return nil, fmt.Errorf("skill `%s` not found", skillName)
}

// ChainCall to call multiple functions in a row
// context is passed to all functions and context["data"] is updated with the response of each function
func (sk *SemanticKernel) ChainCall(context llm.Content, functions ...*Function) (response llm.Content, err error) {
	for _, function := range functions {
		if response, err = function.Call(context); err != nil {
			err = fmt.Errorf("error calling function `%s`: %w", function.Name, err)
			return
		}
		context.Set(response.Value())
	}
	return
}

// func (k *SemanticKernel) ImportSkill(name string) (skill Skill, err error) {
// 	fs, err := fs.Sub(embeddedSkillsDir, "assets/skills")
// 	if err != nil {
// 		return
// 	}

// 	// return k.importSkill(fs, name)
// }

// func (k *SemanticKernel) ImportSkill(name string) (skill Skill, err error) {
// 	fs, err := fs.Sub(embeddedSkillsDir, "assets/skills")
// 	if err != nil {
// 		return
// 	}
// 	// return k.importSkill(fs, name)
// }
