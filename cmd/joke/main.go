package main

import (
	"fmt"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/mfmayer/gosk/pkg/skills/fun"
)

func main() {
	kernel := gosk.NewKernel()
	kernel.RegisterGenerators(gpt.Register)
	kernel.RegisterSkills(fun.Register)
	functions, err := kernel.FindFunctions("fun.joke")
	if err != nil {
		panic(err)
	}
	input := llm.NewContent("dinosaur").With("style", "One-Liner, no question")
	response, err := kernel.Call(input, functions...)
	if err != nil {
		panic(err)
	}
	fmt.Println(response.String())
}
