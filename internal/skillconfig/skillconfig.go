package skillconfig

type Parameter struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DefaultValue string `json:"defaultValue"`
}

// DefaultParameter returns Paramter with name `input`
func DefaultParameter() Parameter {
	return Parameter{
		Name: "input",
	}
}

type Input struct {
	Parameters []Parameter `json:"parameters"`
}

// DefaultInput returns Input with a default parameter generated with `DefaultParameter()`
func DefaultInput() Input {
	return Input{
		Parameters: []Parameter{DefaultParameter()},
	}
}

type Completion struct {
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"top_p"`
	PresencePenalty  float64 `json:"presence_penalty"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
}

func DefaultCompletion() Completion {
	return Completion{
		MaxTokens:        1000,
		Temperature:      0.9,
		TopP:             0.0,
		PresencePenalty:  0.0,
		FrequencyPenalty: 0.0,
	}
}

type SkillConfig struct {
	Schema      int        `json:"schema"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	Completion  Completion `json:"completion"`
	Input       Input      `json:"input"`
}

func DefaultSkillConfig() SkillConfig {
	return SkillConfig{
		Schema:      1,
		Description: "",
		Type:        "completion",
		Completion:  DefaultCompletion(),
		Input:       DefaultInput(),
	}
}
