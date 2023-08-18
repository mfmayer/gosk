package test

import (
	"log"
	"os"
	"testing"
	"text/template"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
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

{{.}}
+++++
`
	template, err := template.New("skprompt").Parse(skprompt)
	if err != nil {
		t.Fatal(err)
	}
	generator, err := gpt.Factory.New(nil)
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
	generator, err := gpt.Factory.New(nil)
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

func TestContentProperties(t *testing.T) {
	generator, err := gpt.Factory.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	skillFunc := func(input llm.Content) (output llm.Content, err error) {
		return generator.Generate(input)
	}

	skprompt := `In the following you are getting a request from {{or .firstName "the user"}} and you are his personal assistant.

THE FOLLOWING FUNCTIONS CAN BE USED AND COMBINED TO FULLFILL THE USERS REQUEST.
[FUNCTIONS]
{
	"GetWeather": {
		"description": "Get the weather forecast for today.", 
		"inputProperties": {
			"location": {
				"description": "The location geocoordinates for which the weather information should be retrieved.",
				"type": "object",
				"properties": {
					"latitude": {
						"description": "The location's geocoordinates latitude.",
						"type": "number"
					},
					"longitude": {
						"description": "The location's geocoordinates longitude.",
						"type": "number"
					}
				}
			}
		}
	},
	"GetUserLocation": {
		"description": "Get the location geocoordinates from {{or .firstName "the user"}}",
		"inputProperties": {}
	}
}
[END FUNCTIONS]

THE FOLLOWING FUNCTIONS HAVE BEEN ALREADY CALLED AND RETURNED THESE RESPONSES:
[
]

PROVIDE A LIST OF FUNCTIONS THAT YOU MUST CALL NEXT IN ORDER GET ALL REMAINING INFOS YOU NEED TO FULLFILL THE REQUEST. USE ONLY PROPERTIES YOU KNOW. USE ONL THE GIVEN FUNCTIONS TO GET PROPERTIES YOU NEED. PROVIDE THE LIST IN THE FOLLOWING FORMAT:
[
	{
		"designation": "Precise designation of the function call's response.",
		"function": "Function name that shall be called",
		"inputProperties": {
			...
		}
	},
	...
]

IF ALL NEEDED INFORMATION IS AVALILABLE PROVIDE AN EMPTY LIST: "[]"

+++++
USER REQUEST: {{.request}}
+++++


`

	template, err := llm.TemplateFromText(skprompt)
	if err != nil {
		t.Fatal(err)
	}

	systemInput := llm.NewContent().
		SetRole(llm.RoleSystem).
		// With("weather", "{\"temperature\": 20, \"wind\": 10, \"rain\": 0, \"clouds\": 0, \"humidity\": 0, \"pressure\": 0, \"condition\": \"sunny\"}").
		// With("weather", "{\"condition\": \"unknown\"}").
		With("firstName", "Max").
		With("botName", "Ida")
		// With("location.latitude", 52.520008).
		// With("location.longitude", 13.404954)
	systemPrompt, err := llm.ExecuteTemplate(template, systemInput)
	if err != nil {
		log.Fatal(err)
	}
	systemInput.Set(systemPrompt).With("request", "Wie wird das Wetter heute?")

	// input := llm.NewContent("Wie wird das Wetter heute?").WithPredecessor(systemInput)
	// input = llm.NewContent("***FUNCTION_CALL:***: GetWeather()").SetRole(llm.RoleAssistant).WithPredecessor(input)

	result, err := skillFunc(systemInput)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
