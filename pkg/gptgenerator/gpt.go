package gptgenerator

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/mfmayer/gopenai"
	"github.com/mfmayer/gosk/pkg/llm"
)

func init() {
	godotenv.Load()
}

// GetOpenAIKey tries to retrieve OpenAI key from "OPENAI_API_KEY" environment variable or .env file in current workgin directory
func getOpenAIKey() (string, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return "", errors.New("openai api key not found")
	}
	return key, nil
}

// Content2Message translates llm.Content into OpenAI Message
func Content2Message(content llm.Content) (msg *gopenai.Message, err error) {
	if content == nil {
		return
	}
	msg = &gopenai.Message{}
	role, roleOK := content.RoleOption()
	if roleOK {
		switch role {
		case llm.RoleSystem:
			msg.Role = gopenai.RoleSystem
		case llm.RoleUser:
			msg.Role = gopenai.RoleUser
		case llm.RoleAssistant:
			msg.Role = gopenai.RoleAssistant
		case llm.RoleFunctionCall:
			msg.Role = gopenai.RoleAssistant
		case llm.RoleFunctionResponse:
			msg.Role = gopenai.RoleFunction
		}
	} else {
		// if no role option is set, use default user role
		msg.Role = gopenai.RoleUser
	}
	if role == llm.RoleFunctionCall {
		msg.FunctionCall = &gopenai.FunctionCall{}
		msg.FunctionCall.Arguments = content.StringData()
		if name, ok := content.NameOption(); ok {
			msg.FunctionCall.Name = name
		} else {
			err = errors.New("content for function call is not designated")
			return
		}
		return
	}
	msg.Content = content.StringData()
	if name, ok := content.NameOption(); ok {
		msg.Name = name
	}
	return
}

// Message2Content translates OpenAI Message into llm.Content
func Message2Content(msg *gopenai.Message) (content llm.Content) {
	if msg == nil {
		return
	}
	name := msg.Name
	contentString := msg.Content
	role := llm.RoleEmpty
	switch msg.Role {
	case gopenai.RoleAssistant:
		role = llm.RoleAssistant
	case gopenai.RoleUser:
		role = llm.RoleUser
	case gopenai.RoleSystem:
		role = llm.RoleSystem
	case gopenai.RoleFunction:
		role = llm.RoleFunctionResponse
	}
	if msg.FunctionCall != nil {
		contentString = msg.FunctionCall.Arguments
		name = msg.FunctionCall.Name
		role = llm.RoleFunctionCall
	}
	content = llm.NewContent(contentString)
	if len(role) > 0 {
		content.WithRoleOption(role)
	}
	if len(name) > 0 {
		content.WithNameOption(name)
	}
	return
}
