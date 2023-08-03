package chat

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
)

//go:embed assets/*
var fsTemplates embed.FS

func New() (*gosk.Skill, error) {
	generator, err := gpt.NewGenerator(gpt.WithModel("gpt-3.5-turbo"))
	if err != nil {
		return nil, err
	}

	skill := &gosk.Skill{
		Name:        "chat",
		Description: "Chat skill",
		Functions:   map[string]*gosk.Function{},
	}
	chatFunction := &gosk.Function{
		Name:        "chatgpt",
		Description: "Chat with GPT-3",
		Parameters: map[string]*gosk.Parameter{
			"data": {
				Description: "Data to be used for chat",
				Required:    true,
			},
			"date": {
				Description: "Today's date",
			},
			"botName": {
				Description: "Name of the bot",
				Default:     "Ida",
			},
			"attitude": {
				Description: "Attitude of the bot",
			},
			"firstName": {
				Description: "First name of the user",
			},
			"language": {
				Description: "Language spoken by the user",
				Required:    true,
				Default:     "english",
			},
		},
	}
	skill.Functions[chatFunction.Name] = chatFunction

	// parse template for system prompt
	template, err := template.ParseFS(fsTemplates, "assets/chatgpt/skprompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed parsing template: %w", err)
	}
	chatFunction.Call = func(input llm.Content) (output llm.Content, err error) {
		// add system prompt to input if not already present
		if input.Predecessor() == nil {
			var promptBuffer bytes.Buffer
			if err = template.Execute(&promptBuffer, input); err != nil {
				return
			}
			systemInput := llm.NewContent(promptBuffer.String()).SetRole(llm.RoleSystem)
			input.WithPredecessor(systemInput)
		}
		response, err := generator.GenerateResponse(input)
		return response.WithPredecessor(input), err
	}
	return skill, nil
}
