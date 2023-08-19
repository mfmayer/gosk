package planner

import (
	"embed"
	"fmt"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
)

//go:embed assets/*
var fsTemplates embed.FS

func New() (*gosk.Skill, error) {
	generator, err := gpt.Factory.New(nil)
	if err != nil {
		return nil, err
	}

	skill := &gosk.Skill{
		Name:        "planner",
		Description: "Planner skill",
		Functions:   map[string]*gosk.Function{},
		Plannable:   false,
	}
	planFunction := &gosk.Function{
		Name:        "planner",
		Description: "Skill Planner",
		Plannable:   false,
		InputProperties: map[string]*gosk.Parameter{
			"skills": {
				Description: "Available skills",
				Required:    true,
			},
		},
	}
	skill.Functions[planFunction.Name] = planFunction

	// parse template for system prompt
	template, err := llm.TemplateFromFS(fsTemplates, "assets/chatgpt/skprompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed parsing template: %w", err)
	}
	planFunction.Call = func(input llm.Content) (llm.Content, error) {
		// add system prompt to input if not already present
		if input.Predecessor() == nil {
			systemPrompt, err := llm.ExecuteTemplate(template, input)
			if err != nil {
				return nil, err
			}
			systemInput := llm.NewContent(systemPrompt).SetRole(llm.RoleSystem)
			input.WithPredecessor(systemInput)
		}
		response, err := generator.Generate(input)
		return response.WithPredecessor(input), err
	}
	return skill, nil
}
