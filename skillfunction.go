package gosk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"text/template"

	"github.com/mfmayer/gosk/pkg/llm"
)

type Type string

const (
	TypeString  Type = "string"
	TypeNumber  Type = "number"
	TypeInteger Type = "integer"
	TypeObject  Type = "object"
	TypeArray   Type = "array"
	TypeBoolean Type = "boolean"
	TypeNull    Type = "null"
)

// Parameter defines a function's parameter with its name, description and type
type Parameter struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description"`
	Type        Type        `json:"type,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	//TODO: Add additional potentially valuable definitions like min, max, etc
}

// Function defines and describes a skill's function with its input properties and its actual function call
type Function struct {
	// Name of the SkillFunction
	Name string `json:"name,omitempty"`
	// Description what the SkillFunction is doing
	Description string `json:"description"`
	// Plannable indicates whether the skill function can be planned by the semantic kernel
	Plannable bool `json:"plannable,omitempty"`
	// InputProperties map whose keys are the input property names and whose values are the input property definitions
	InputProperties map[string]*Parameter `json:"inputProperties"`
	// call holds the function that is executed when the skill function is called
	Call func(input llm.Content) (output llm.Content, err error) `json:"-"`
}

// functionConfig is used to unmarshal function configuration into it
type functionConfig struct {
	*Function
	Generator string `json:"generator"`
}

type createSemanticFunctionsOptionProperties struct {
	createSemanticFunctions map[string]func(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(input llm.Content) (response llm.Content, err error))
}

type createSemanticFunctionsOption func(properties *createSemanticFunctionsOptionProperties)

// WithCustomCallForFunc allows to create selectively custom semantic function calls while parsing multiple semantic functions with ParseSemanticFunctionsFromFS
func WithCustomCallForFunc(funcName string, createSemanticFunctionCall func(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(input llm.Content) (response llm.Content, err error))) (option createSemanticFunctionsOption) {
	option = func(properties *createSemanticFunctionsOptionProperties) {
		properties.createSemanticFunctions[funcName] = createSemanticFunctionCall
	}
	return
}

func ParseSemanticFunctionsFromFS(fsys fs.FS, generators map[string]llm.Generator, options ...createSemanticFunctionsOption) (functions map[string]*Function, err error) {
	optionProperties := createSemanticFunctionsOptionProperties{
		createSemanticFunctions: map[string]func(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(input llm.Content) (response llm.Content, err error)){},
	}
	for _, option := range options {
		option(&optionProperties)
	}
	functions = map[string]*Function{}
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
		subFS, subErr := fs.Sub(fsys, d.Name())
		if subErr != nil {
			// skip subdirectory if it is not possible to create subFS
			continue
		}
		// functionName
		functionName := strings.ToLower(d.Name())
		var function *Function
		var parseFunctionErr error
		// check for custom option
		if custemCreateSemanticFunctionCall, ok := optionProperties.createSemanticFunctions[functionName]; ok {
			// create function with given option
			function, parseFunctionErr = ParseSemanticFunctionFromFS(subFS, generators, WithCustomCall(custemCreateSemanticFunctionCall))
		} else {
			// create default function
			function, parseFunctionErr = ParseSemanticFunctionFromFS(subFS, generators)
		}
		if parseFunctionErr != nil {
			if !errors.Is(parseFunctionErr, fs.ErrNotExist) {
				// if function config file doesn't exist ignore the error and
				// go to next subdirectory - otherwise join the error for returning it
				err = errors.Join(err, parseFunctionErr)
			}
			// skip subdirectory if it is not possible to create function
			continue
		}
		// add function to skill with its directory name as key
		functions[functionName] = function
	}
	return
}

type parseSemanticFunctionFromFSOptionProperties struct {
	createSemanticFunction func(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(input llm.Content) (response llm.Content, err error))
}

type parseSemanticFunctionFromFSOption func(properties *parseSemanticFunctionFromFSOptionProperties)

// WithCustomCall allows to create a custom semantic function call while parsing a semantic function with ParseSemanticFunctionFromFS
func WithCustomCall(createSemanticFunctionCall func(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(input llm.Content) (response llm.Content, err error))) (option parseSemanticFunctionFromFSOption) {
	option = func(properties *parseSemanticFunctionFromFSOptionProperties) {
		properties.createSemanticFunction = createSemanticFunctionCall
	}
	return
}

// ParseFunctionFromFS finds "config.json" with function comfiguration.
// Prompt templates will be created from "*.tmpl" files with at least "skprompt.tmpl" is needed
func ParseSemanticFunctionFromFS(fsys fs.FS, generators map[string]llm.Generator, options ...parseSemanticFunctionFromFSOption) (function *Function, err error) {
	optionProperties := parseSemanticFunctionFromFSOptionProperties{
		createSemanticFunction: NewDefaultSemanticFunctionCall,
	}
	for _, option := range options {
		option(&optionProperties)
	}
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
	var functionConfig functionConfig
	err = json.Unmarshal(data, &functionConfig)
	function = functionConfig.Function
	if err != nil {
		err = fmt.Errorf("unmarshalling `config.json` failed: %w", err)
		return
	}

	// check if supported generator is avalilable
	generator, ok := generators[functionConfig.Generator]
	if !ok || generator == nil {
		err = fmt.Errorf("generator \"%s\" not found for function \"%s\"", functionConfig.Generator, function.Name)
		return
	}

	// get template
	template, err := llm.TemplateFromFS(fsys, "*.tmpl")

	// create function call
	function.Call = optionProperties.createSemanticFunction(template, generator)
	return
}

// NewDefaultSemanticFunctionCall creates a new semantic skill function with a prompt template and a generator
func NewDefaultSemanticFunctionCall(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(input llm.Content) (response llm.Content, err error)) {
	if promptTemplate == nil {
		return
	}
	skillFunc = func(input llm.Content) (output llm.Content, err error) {
		var promptBuffer bytes.Buffer
		if err = promptTemplate.Execute(&promptBuffer, input); err != nil {
			return
		}
		input.Set(promptBuffer.String())
		return generator.Generate(input)
	}
	return
}
