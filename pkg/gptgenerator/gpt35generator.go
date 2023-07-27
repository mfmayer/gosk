package gptgenerator

import (
	"errors"

	"github.com/mfmayer/gopenai"
	"github.com/mfmayer/gosk/pkg/llm"
)

// GPT35Generator represents the OpenAI GPT3.5 Model and implements the llm.Generator interface
type GPT35Generator struct {
	chatClient *gopenai.ChatClient
}

// NewGPT35Generator creates new GPT35Generator
func NewGPT35Generator() (generator *GPT35Generator, err error) {
	key, err := getOpenAIKey()
	if err != nil {
		return
	}
	chatClient := gopenai.NewChatClient(key)
	generator = &GPT35Generator{
		chatClient: chatClient,
	}
	return
}

// GenerateResponse to get response from the model
func (gpt *GPT35Generator) GenerateResponse(input llm.Content) (response llm.Content, err error) {
	if gpt.chatClient == nil {
		err = errors.New("missing model client")
		return
	}
	inputSlice := []llm.Content{input}

	predecessor, ok := input.Predecessor()
	for {
		if !ok {
			break
		}
		inputSlice = append(inputSlice, predecessor)
		predecessor, ok = predecessor.Predecessor()
	}

	chatPrompt := gopenai.ChatPrompt{
		Model:    "gpt-3.5-turbo",
		Messages: make([]*gopenai.Message, 0, len(inputSlice)),
	}
	for i := len(inputSlice) - 1; i >= 0; i-- {
		inputElement := inputSlice[i]
		if msg, err := Content2Message(inputElement); err == nil {
			chatPrompt.Messages = append(chatPrompt.Messages, msg)
		}
	}
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
