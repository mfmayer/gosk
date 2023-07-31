package test

import (
	"testing"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/mfmayer/gosk/pkg/skills/fun"
	"github.com/mfmayer/gosk/pkg/skills/writer"
)

func TestKernel(t *testing.T) {
	kernel := gosk.NewKernel()
	err := kernel.CreateAndAddSkills(fun.New, writer.New)
	if err != nil {
		t.Fatal(err)
	}
	functions, err := kernel.FindFunctions("fun.joke", "writer.translate")
	if err != nil {
		t.Fatal(err)
	}
	result, err := kernel.ChainCall(llm.NewContent("flowers").With("language", "german"), functions...)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
