package gosk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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
	// DefaultValue interface{} `json:"defaultValue,omitempty"` //TODO: to be implemented
	//TODO: Add additional potentially valuable definitions like min, max, etc
}

// Function defines and describes a skill's function with its parameters and its actual function call
type Function struct {
	// Name of the SkillFunction
	Name string `json:"name,omitempty"`
	// Description what the SkillFunction is doing
	Description string `json:"description"`
	// Plannable indicates whether the skill function can be planned by the semantic kernel
	Plannable bool `json:"plannable,omitempty"`
	// Parameters map whose keys are the parameters name and values their definition
	Parameters map[string]*Parameter `json:"parameters"`
	// call holds the function that is executed when the skill function is called
	Call func(input llm.Content) (output llm.Content, err error) `json:"-"`
}

type functionConfig struct {
	*Function
	Generator string `json:"generator"`
}

// ParseFunctionFromFS finds "config.json" with function comfiguration.
// Prompt templates will be created from "*.tmpl" files with at least "skprompt.tmpl" is needed
func ParseSemanticFunctionFromFS(fsys fs.FS, generators llm.GeneratorMap) (function *Function, err error) {
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
	// resolve parameter names
	// TODO: check if this is needed
	// for paramName, param := range function.Parameters {
	// 	param.Name = paramName
	// }

	// check if supported generator is avalilable
	generator, ok := generators[functionConfig.Generator]
	if !ok || generator == nil {
		err = fmt.Errorf("generator \"%s\" not found for function \"%s\"", functionConfig.Generator, function.Name)
		return
	}

	// get template
	template, err := template.ParseFS(fsys, "*.tmpl")
	if err != nil {
		err = fmt.Errorf("error parsing templates: %w", err)
		return
	}
	promptTemplate := template.Lookup("skprompt.tmpl")
	if promptTemplate == nil {
		err = fmt.Errorf("\"skprompt.tmpl\" not found")
		return
	}

	// create function call
	function.Call = NewSemanticFunctionCall(promptTemplate, generator)
	return
}

// NewSemanticFunctionCall creates a new semantic skill function with a prompt template and a generator
func NewSemanticFunctionCall(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(parameters llm.Content) (response llm.Content, err error)) {
	skillFunc = func(input llm.Content) (output llm.Content, err error) {
		// FIXME: parameter availability must be checked
		var promptBuffer bytes.Buffer
		if err = promptTemplate.Execute(&promptBuffer, input); err != nil {
			return
		}
		input.WithData(promptBuffer.String())
		return generator.GenerateResponse(input)
	}
	return
}
