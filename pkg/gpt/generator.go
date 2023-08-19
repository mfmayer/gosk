package gpt

import (
	"errors"

	"github.com/mfmayer/gopenai"
	"github.com/mfmayer/gosk/pkg/llm"
)

var Factory *gptFactory

type gptFactory struct{}

func (f *gptFactory) TypeID() string {
	return "gpt"
}

func (f *gptFactory) New(config llm.GeneratorConfigData) (generator llm.Generator, err error) {
	key, err := getOpenAIKey()
	if err != nil {
		return
	}
	chatClient := gopenai.NewChatClient(key)
	// generatorConfig := gopenai.ChatPromptConfig{}
	// config.Convert(&generatorConfig)
	gptGenerator := &Generator{
		//model:      "gpt-3.5-turbo", // FIXME: should be set by config
		config:     &gopenai.ChatPromptConfig{},
		chatClient: chatClient,
	}
	config.Convert(gptGenerator.config)
	generator = gptGenerator
	return
}

// GPT35Generator represents the OpenAI GPT3.5 Model and implements the llm.Generator interface
type Generator struct {
	config     *gopenai.ChatPromptConfig
	chatClient *gopenai.ChatClient
}

// GenerateResponse to get response from the model
func (gpt *Generator) Generate(input llm.Content) (response llm.Content, err error) {
	if gpt.chatClient == nil {
		err = errors.New("missing model client")
		return
	}

	// get all predecessors and append them to input slice
	inputSlice := []llm.Content{input}
	predecessor := input.Predecessor()
	for {
		if predecessor == nil {
			break
		}
		inputSlice = append(inputSlice, predecessor)
		predecessor = predecessor.Predecessor()
	}

	// create chat prompt
	chatPrompt := gopenai.ChatPrompt{
		ChatPromptConfig: gpt.config,
		Messages:         make([]*gopenai.Message, 0, len(inputSlice)),
	}
	// iterate over input slice in reverse order to get the correct order of messages
	for i := len(inputSlice) - 1; i >= 0; i-- {
		inputElement := inputSlice[i]
		if msg, err := Content2Message(inputElement); err == nil {
			chatPrompt.Messages = append(chatPrompt.Messages, msg)
		}
	}
	// get response
	completion, err := gpt.chatClient.GetChatCompletion(&chatPrompt)
	if err != nil {
		return
	}
	if completion.Error != nil {
		err = errors.New(completion.Error.Message)
		return
	}
	if len(completion.Choices) <= 0 {
		err = errors.New("no resposne available")
	}
	response = Message2Content(&completion.Choices[0].Message)
	return
}
