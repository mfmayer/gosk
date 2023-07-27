package test

import (
	"testing"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gptgenerator"
	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/mfmayer/gosk/pkg/skills"
)

func TestKernel(t *testing.T) {
	kernel := gosk.NewKernel()
	generator, err := gptgenerator.NewGPT35Generator()
	if err != nil {
		t.Fatal(err)
	}
	generators := map[string]llm.Generator{
		"gpt35": generator,
	}
	skills, err := skills.CreateSemanticSkills(generators, "fun", "writer")
	if err != nil {
		t.Fatal(err)
	}
	kernel.AddSkillsMap(skills)
	functions, err := kernel.FindFunctions("fun.joke", "writer.translate")
	if err != nil {
		t.Fatal(err)
	}
	result, err := kernel.ChainCall(llm.NewContent("dinosaurs").With("language", "german"), functions...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
