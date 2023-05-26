package gosk

import (
	"bytes"
	"embed"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/mfmayer/gopenai"
	"github.com/mfmayer/gosk/internal/skillconfig"
	"github.com/mfmayer/gosk/utils"
)

//go:embed assets/skills/*
var embeddedSkillsDir embed.FS

// SemanticKernel
type SemanticKernel struct {
	chatClient *gopenai.ChatClient
}

type newKernelOption func(*newKernelOptions)

type newKernelOptions struct {
	openAIKey string
}

// WithOpenAIKey to use this OpenAI key when creating a new semantic kernel, otherwise it's tried to get the key from "OPENAI_API_KEY" environment variable or .env file in current working directory
func WithOpenAIKey(key string) newKernelOption {
	return func(opt *newKernelOptions) {
		opt.openAIKey = key
	}
}

// NewKernel creates new kernel and tries to retrieve the OpenAI key from "OPENAI_API_KEY" environment variable or .env file in current working directory
func NewKernel(opts ...newKernelOption) (kernel *SemanticKernel, err error) {
	options := &newKernelOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.openAIKey == "" {
		options.openAIKey, err = utils.GetOpenAIKey()
		if err != nil {
			return
		}
	}
	cClient := gopenai.NewChatClient(options.openAIKey)
	kernel = &SemanticKernel{
		chatClient: cClient,
	}
	return
}

func (k *SemanticKernel) ImportSkill(name string) (skill Skill, err error) {
	fs, err := fs.Sub(embeddedSkillsDir, "assets/skills")
	if err != nil {
		return
	}
	return k.importSkill(fs, name)
}

func (k *SemanticKernel) importSkill(fsys fs.FS, skillName string) (skill Skill, err error) {
	skill = Skill{}
	err = fs.WalkDir(fsys, skillName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != skillName {
			skPromptPath := filepath.Join(path, "skprompt.txt")
			skConfigPath := filepath.Join(path, "config.json")
			if !utils.FilesExist(fsys, skPromptPath) {
				// ignore path when there is no prompt (config.json is optional)
				return nil
			}
			// read skill prompt template
			skprompt, err := template.ParseFS(fsys, skPromptPath)
			_ = skprompt
			if err != nil {
				return err
			}
			// create default config and read optional skill config file
			sConfig := skillconfig.DefaultSkillConfig()
			jsonFile, err := fsys.Open(skConfigPath)
			if err == nil {
				defer jsonFile.Close()
				sConfigBytes, _ := ioutil.ReadAll(jsonFile)
				_ = json.Unmarshal(sConfigBytes, &sConfig)
				// if err != nil {
				// 	return err
				// }
			}
			// use path base name as skill function name
			skillFunctionName := filepath.Base(path)
			skill[skillFunctionName] = k.createSkillFunction(skprompt, sConfig)
		}
		return nil
	})
	return
}

func (k *SemanticKernel) createSkillFunction(template *template.Template, config skillconfig.SkillConfig) (skillFunc SkillFunction) {
	paramMap := map[string]*string{}
	paramArray := []*string{}
	for _, p := range config.Input.Parameters {
		param := p.DefaultValue
		paramMap[p.Name] = &param
		paramArray = append(paramArray, &param)
	}

	skillFunc = func(parameters ...string) (response string, err error) {
		for i, param := range parameters {
			if i < len(paramArray) {
				*paramArray[i] = param
			}
		}

		var promptBuffer bytes.Buffer
		template.Execute(&promptBuffer, paramMap)

		chatPrompt := gopenai.ChatPrompt{
			Model: "gpt-3.5-turbo",
			Messages: []*gopenai.Message{
				{
					Role:    "user",
					Content: promptBuffer.String(),
				},
			},
		}
		var completion *gopenai.ChatCompletion
		completion, err = k.chatClient.GetChatCompletion(&chatPrompt)
		if err == nil {
			response = completion.Choices[0].Message.Content
		}
		return
	}
	return
}
