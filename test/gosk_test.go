package test

import (
	"fmt"
	"testing"

	"github.com/mfmayer/gosk"
)

func TestKernel(t *testing.T) {
	kernel, err := gosk.NewKernel()
	if err != nil {
		t.Error(err)
	}
	t.Logf("hello kernel: %v", kernel)
}

func TestSkillImport(t *testing.T) {
	kernel, err := gosk.NewKernel()
	if err != nil {
		t.Fatal(err)
	}
	skill, err := kernel.ImportSkill("FunSkill")
	if err != nil {
		t.Fatal(err)
	}
	sf := skill["Joke"]
	response, err := sf("Engineer", "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", response)
}
