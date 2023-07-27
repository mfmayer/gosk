package test

import (
	"testing"

	"github.com/mfmayer/gosk/pkg/gptgenerator"
	"github.com/mfmayer/gosk/pkg/llm"
)

func TestGenerator(t *testing.T) {
	var generator llm.Generator
	var err error
	generator, err = gptgenerator.NewGPT35Generator()
	if err != nil {
		t.Fatal(err)
	}
	input := llm.NewContent("Du heißt Ida und bist ein persönlicher Assistent.").WithRoleOption(llm.RoleSystem)
	input = llm.NewContent("Hallo!").WithRoleOption(llm.RoleUser).WithPredecessor(input)
	input = llm.NewContent("Hallo! Wie kann ich Ihnen helfen?").WithRoleOption(llm.RoleAssistant).WithPredecessor(input)
	input = llm.NewContent("Wie heisst du?").WithRoleOption(llm.RoleUser).WithPredecessor(input)

	response, err := generator.GenerateResponse(input)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response)
}
