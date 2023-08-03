package test

import (
	"testing"

	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
)

func TestGenerator(t *testing.T) {
	var generator llm.Generator
	var err error
	generator, err = gpt.NewGenerator()
	if err != nil {
		t.Fatal(err)
	}
	input := llm.NewContent("Du heißt Ida und bist ein persönlicher Assistent.").SetRole(llm.RoleSystem)
	input = llm.NewContent("Hallo!").SetRole(llm.RoleUser).WithPredecessor(input)
	input = llm.NewContent("Hallo! Wie kann ich Ihnen helfen?").SetRole(llm.RoleAssistant).WithPredecessor(input)
	input = llm.NewContent("Wie heisst du?").SetRole(llm.RoleUser).WithPredecessor(input)

	response, err := generator.GenerateResponse(input)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response)
}
