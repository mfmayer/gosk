package test

import (
	"os"
	"testing"
	"text/template"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gptgenerator"
	"github.com/mfmayer/gosk/pkg/llm"
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
	generator, err := gptgenerator.NewGPT35Generator()
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
	// path, err := os.Getwd()
	// if err != nil {
	// 	t.Fatal("Fehler beim Abrufen des Arbeitsverzeichnisses:", err)
	// }
	// t.Log(path)
	generator, err := gptgenerator.NewGPT35Generator()
	if err != nil {
		t.Fatal(err)
	}
	generators := map[string]llm.Generator{
		"gpt35": generator,
	}

	fsys := os.DirFS("../pkg/skills/assets/fun/joke")
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
	generator, err := gptgenerator.NewGPT35Generator()
	if err != nil {
		t.Fatal(err)
	}
	generators := map[string]llm.Generator{
		"gpt35": generator,
	}

	fsys := os.DirFS("../pkg/skills/assets/fun")
	skill, err := gosk.ParseSemanticSkillFromFS(fsys, generators)
	if err != nil {
		t.Fatal(err)
	}
	result, err := skill.Call("joke", llm.NewContent("dinosaurs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}

func TestParseSemanticSkills(t *testing.T) {
	generator, err := gptgenerator.NewGPT35Generator()
	if err != nil {
		t.Fatal(err)
	}
	generators := map[string]llm.Generator{
		"gpt35": generator,
	}
	fsys := os.DirFS("../pkg/skills/assets")
	skills, err := gosk.ParseSemanticSkillsFromFS(fsys, generators)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(skills)
}
