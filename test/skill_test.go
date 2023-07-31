package test

import (
	"log"
	"os"
	"testing"
	"text/template"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/mfmayer/gosk/pkg/skills/fun"
)

func TestNewSemanticFunctionCall(t *testing.T) {
	skprompt := `WRITE EXACTLY ONE JOKE or HUMOROUS STORY ABOUT THE TOPIC BELOW

JOKE MUST BE:
- G RATED
- WORKPLACE/FAMILY SAFE
NO SEXISM, RACISM OR OTHER BIAS/BIGOTRY

BE CREATIVE AND FUNNY. I WANT TO LAUGH.
{{or .style ""}}
+++++

{{.data}}
+++++
`
	template, err := template.New("skprompt").Parse(skprompt)
	if err != nil {
		t.Fatal(err)
	}
	generator, err := gpt.NewGenerator()
	if err != nil {
		t.Fatal(err)
	}
	skillFunc := gosk.NewSemanticFunctionCall(template, generator)
	result, err := skillFunc(llm.NewContent("dinosaurs").With("style", "as a shortstory"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}

func TestParseSemanticFunctionFromFS(t *testing.T) {
	generator, err := gpt.NewGenerator()
	if err != nil {
		t.Fatal(err)
	}
	generators := map[string]llm.Generator{
		"gpt-3.5-turbo": generator,
	}
	fsys := os.DirFS("../pkg/skills/fun/assets/joke")
	function, err := gosk.ParseSemanticFunctionFromFS(fsys, generators)
	if err != nil {
		t.Fatal(err)
	}
	result, err := function.Call(llm.NewContent("dinosaurs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}

func TestParseSemanticSkillFromFS(t *testing.T) {
	getGenerators := func(configs map[string]llm.GeneratorConfig) (generators llm.GeneratorMap, err error) {
		generators = llm.GeneratorMap{}
		for k, v := range configs {
			generator, err := gpt.NewGenerator(gpt.WithConfig(v))
			if err != nil {
				return nil, err
			}
			generators[k] = generator
		}
		return
	}
	fsys := os.DirFS("../pkg/skills/fun/assets")
	skill, err := gosk.ParseSemanticSkillFromFS(fsys, getGenerators)
	if err != nil {
		t.Fatal(err)
	}
	result, err := skill.Call("joke", llm.NewContent("dinosaurs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}

func TestFunSkill(t *testing.T) {
	skill, err := fun.New()
	if err != nil {
		log.Fatal(err)
	}
	result, err := skill.Call("joke", llm.NewContent("dinosaurs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
