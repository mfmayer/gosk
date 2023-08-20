package chat

import (
	"embed"
	"io/fs"
	"text/template"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/llm"
)

//go:embed assets/*
var fsAssets embed.FS

func Register(generatorFactories llm.NewGeneratorFuncMap) (skill *gosk.Skill, err error) {
	createChatFunction := func(promptTemplate *template.Template, generator llm.Generator) (skillFunc func(input llm.Content) (response llm.Content, err error)) {
		skillFunc = func(input llm.Content) (llm.Content, error) {
			// add system at the beginning of the conversation (when there is no input's predecessor)
			if input.Predecessor() == nil {
				systemPrompt, err := llm.ExecuteTemplate(promptTemplate, input)
				if err != nil {
					return nil, err
				}
				systemInput := llm.NewContent(systemPrompt).SetRole(llm.RoleSystem)
				input.WithPredecessor(systemInput)
			}
			response, err := generator.Generate(input)
			return response.WithPredecessor(input), err
		}
		return
	}

	subFS, err := fs.Sub(fsAssets, "assets")
	if err != nil {
		return
	}
	skill, err = gosk.ParseSemanticSkillFromFS(subFS, generatorFactories, gosk.WithCustomCallForFunc("chatgpt", createChatFunction))
	if err != nil {
		return
	}
	return skill, nil
}
