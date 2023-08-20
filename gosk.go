package gosk

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mfmayer/gosk/pkg/llm"
)

var (
	// ErrMissingParameter is returned when a required parameter is missing
	ErrGeneratorAlreadyRegistered = errors.New("generator already registered")
	ErrMissingParameter           = errors.New("missing parameter")
	ErrSkillNotFound              = errors.New("skill not found")
	ErrFunctionNotFound           = errors.New("function not found")
)

// SemanticKernel
type SemanticKernel struct {
	registeredGenerators llm.NewGeneratorFuncMap
	skills               map[string]*Skill
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
		registeredGenerators: llm.NewGeneratorFuncMap{},
		skills:               map[string]*Skill{},
	}
	return kernel
}

// RegisterGenerators registers new generators with their registaration functions and make them available to skills
func (sk *SemanticKernel) RegisterGenerators(registrationFuncs ...llm.RegistrationFunc) (err error) {
	for _, factory := range registrationFuncs {
		typeID, newGeneratorFunc := factory()
		if _, exists := sk.registeredGenerators[typeID]; exists {
			err = errors.Join(err, fmt.Errorf("%w: %s", ErrGeneratorAlreadyRegistered, typeID))
		}
		sk.registeredGenerators[typeID] = newGeneratorFunc
	}
	return
}

// RegisterSkills registers new skills with their registration functions and adds them to the kernel with their individual names
func (sk *SemanticKernel) RegisterSkills(registrationFuncs ...SkillRegistrationFunc) (err error) {
	for _, registrationFunc := range registrationFuncs {
		skill, registrationErr := registrationFunc(sk.registeredGenerators)
		if registrationErr != nil {
			err = errors.Join(err, fmt.Errorf("error registering %s: %w", skill, registrationErr))
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

// Call one or more functions in a row.
// The given input (incl. all its properties) is passed to each function after it has been
// updated with the previous function's response value.
func (sk *SemanticKernel) Call(input llm.Content, functions ...*Function) (response llm.Content, err error) {
	initialValue := input.Value()
	for _, function := range functions {
		// if response, err = function.Call(context); err != nil {
		if response, err = sk.call(input, function); err != nil {
			err = fmt.Errorf("error calling function `%s`: %w", function.Name, err)
			return
		}
		input.Set(response.Value())
	}
	// restore initial input value
	input.Set(initialValue)
	return
}

// call given function with given input.
func (sk *SemanticKernel) call(input llm.Content, function *Function) (response llm.Content, err error) {
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
	return
}
