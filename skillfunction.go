package gosk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"text/template"

	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/tidwall/gjson"
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

// FunctionParameter defines a function's parameter with its name, description and type
type FunctionParameter struct {
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
	// Parameters map whose keys are the parameters name and values their definition
	Parameters map[string]*FunctionParameter `json:"parameters"`
	// call holds the function that is executed when the skill function is called
	Call func(input llm.Content) (output llm.Content, err error) `json:"-"`
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
	var f Function
	err = json.Unmarshal(data, &f)
	function = &f
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
	generatorConfig, ok := gjson.Get(string(data), "generator").Value().(map[string]interface{})
	if !ok {
		err = fmt.Errorf("no valid `generator` config found in `config.json`")
		return
	}
	var generator llm.Generator
	if supportedGenerators, ok := generatorConfig["name"]; ok {
		if supportedGeneratorsString, ok := supportedGenerators.(string); ok {
			generatorList := strings.Split(supportedGeneratorsString, ",")
			for i, generator := range generatorList {
				generatorList[i] = strings.TrimSpace(generator)
			}
			generator, _ = generators.FindAny(generatorList...)
		}
	}
	if generator == nil {
		err = fmt.Errorf("no supporting generator found")
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
