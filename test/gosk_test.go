package test

import (
	"testing"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/mfmayer/gosk/pkg/skills/fun"
	"github.com/mfmayer/gosk/pkg/skills/writer"
)

func TestKernel(t *testing.T) {
	kernel := gosk.NewKernel()
	kernel.RegisterGenerators(gpt.Register)
	err := kernel.RegisterSkills(fun.Register, writer.Register)
	if err != nil {
		t.Fatal(err)
	}

	functions, err := kernel.FindFunctions("fun.joke", "writer.translate")
	if err != nil {
		t.Fatal(err)
	}
	result, err := kernel.Call(llm.NewContent("flowers").With("language", "german"), functions...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result.String())
}

func TestFunSkill(t *testing.T) {
	kernel := gosk.NewKernel()
	kernel.RegisterGenerators(gpt.Register)
	err := kernel.RegisterSkills(fun.Register)
	if err != nil {
		t.Fatal(err)
	}
	functions, err := kernel.FindFunctions("fun.joke")
	if err != nil {
		t.Fatal(err)
	}
	result, err := kernel.Call(llm.NewContent("dinosaur").
		With("style", "AS A ONE-LINER."),
		functions...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result.String())
}
