package gpt

import (
	"errors"

	"github.com/mfmayer/gopenai"
	"github.com/mfmayer/gosk/pkg/llm"
)

// GPT35Generator represents the OpenAI GPT3.5 Model and implements the llm.Generator interface
type Generator struct {
	model      string
	chatClient *gopenai.ChatClient
}

type optionData struct {
	model  string
	config llm.GeneratorConfig
}

type optionFunc func(*optionData)

// WithModel to set the model
func WithModel(model string) optionFunc {
	return func(o *optionData) {
		o.model = model
	}
}

func WithConfig(config llm.GeneratorConfig) optionFunc {
	return func(o *optionData) {
		o.config = config
	}
}

// NewGPT35Generator creates new GPT35Generator
func NewGenerator(option ...optionFunc) (generator *Generator, err error) {
	key, err := getOpenAIKey()
	if err != nil {
		return
	}
	optionData := optionData{
		model: "gpt-3.5-turbo",
	}
	for _, opt := range option {
		opt(&optionData)
	}
	chatClient := gopenai.NewChatClient(key)
	generator = &Generator{
		model:      optionData.model,
		chatClient: chatClient,
	}
	return
}

func (gpt *Generator) Name() string {
	return gpt.model
}

// GenerateResponse to get response from the model
func (gpt *Generator) GenerateResponse(input llm.Content) (response llm.Content, err error) {
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
		Model:    gpt.model,
		Messages: make([]*gopenai.Message, 0, len(inputSlice)),
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
