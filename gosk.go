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
	ErrSkillNotFound    = errors.New("skill not found")
	ErrFunctionNotFound = errors.New("function not found")
)

// SemanticKernel
type SemanticKernel struct {
	generatorFactories llm.GeneratorFactoryMap
	skills             map[string]*Skill
}

type newKernelOption func(*newKernelOptions)

type newKernelOptions struct {
}

// NewKernel creates new kernel and tries to retrieve the OpenAI key from "OPENAI_API_KEY" environment variable or .env file in current working directory
func NewKernel(opts ...newKernelOption) *SemanticKernel {
	options := &newKernelOptions{}
	for _, opt := range opts {
		opt(options)
	}

	kernel := &SemanticKernel{
		generatorFactories: llm.GeneratorFactoryMap{},
		skills:             map[string]*Skill{},
	}
	return kernel
}

func (sk *SemanticKernel) RegisterGeneratorFactories(factories ...llm.GeneratorFactory) {
	for _, factory := range factories {
		sk.generatorFactories[factory.TypeID()] = factory
	}
}

// RegisterSkills creates new skills with their factories and adds them to the kernel with their individual names
func (sk *SemanticKernel) RegisterSkills(skillFactories ...SkillFactoryFunc) (err error) {
	for _, skillFactory := range skillFactories {
		skill, newSkillErr := skillFactory(sk.generatorFactories)
		if newSkillErr != nil {
			err = errors.Join(err, fmt.Errorf("error registering %s: %w", skill, newSkillErr))
			continue
		}
		err = errors.Join(err, sk.addSkill(skill.Name, skill))
	}
	return
}

// AddSkills adds already initialized skills to the kernel with their individual names
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
		for parameterName, parameter := range function.InputProperties {
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
func (sk *SemanticKernel) FindSkill(skillName string) (skill *Skill, err error) {
	if skill, ok := sk.skills[skillName]; ok {
		return skill, nil
	}
	return nil, ErrSkillNotFound
}

// FindFunction finds a function in a skill by name and returns it or an error if not found
func (sk *SemanticKernel) FindFunction(skillName string, skillFunction string) (function *Function, err error) {
	skill, err := sk.FindSkill(skillName)
	if err != nil {
		return
	}
	if function, ok := skill.Functions[skillFunction]; ok {
		return function, nil
	}
	return nil, ErrFunctionNotFound
}

// FindFunctions finds functions with path notation (`skillName.functionName`) and returns them or an error if any function is not found
func (sk *SemanticKernel) FindFunctions(functionPaths ...string) (functions []*Function, err error) {
	if len(functionPaths) == 0 {
		return nil, fmt.Errorf("%w: missing path", ErrFunctionNotFound)
	}
	functions = make([]*Function, 0, len(functionPaths))
	for _, fp := range functionPaths {
		fps := strings.Split(fp, ".")
		if len(fps) != 2 {
			err = errors.Join(err, fmt.Errorf("%w: path `%s` invalid", ErrFunctionNotFound, fp))
			continue
		}
		function, findErr := sk.FindFunction(fps[0], fps[1])
		if findErr == nil {
			functions = append(functions, function)
		} else {
			err = errors.Join(err, fmt.Errorf("%w: %s", findErr, fp))
		}
	}
	return
}

// CallWithName as shortcut to SemanticKernel.FindFunction and SemanticKernel.Call
func (sk *SemanticKernel) CallWithName(input llm.Content, skillName string, skillFunction string) (response llm.Content, err error) {
	function, err := sk.FindFunction(skillName, skillFunction)
	if err != nil {
		return
	}
	return sk.Call(input, function)
}

// ChainCall to call multiple functions in a row
// context is passed to all functions and context value is updated with the response of each function
func (sk *SemanticKernel) ChainCall(context llm.Content, functions ...*Function) (response llm.Content, err error) {
	for _, function := range functions {
		// if response, err = function.Call(context); err != nil {
		if response, err = sk.Call(context, function); err != nil {
			err = fmt.Errorf("error calling function `%s`: %w", function.Name, err)
			return
		}
		context.Set(response.Value())
	}
	return
}

// Call fiven function with given input. Links the input as predecessor to the response.
func (sk *SemanticKernel) Call(input llm.Content, function *Function) (response llm.Content, err error) {
	if function == nil {
		err = errors.New("function is nil")
		return
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
				err = errors.Join(err, fmt.Errorf("%w: `%s` (function: %s)", ErrMissingParameter, parameter.Name, function.Name))
			}
		}
	}
	if err != nil {
		return nil, err
	}
	// Call function
	response, err = function.Call(input)
	if response != nil {
		// if valid add input as predecessor
		response.WithPredecessor(input)
	}
	return
}
