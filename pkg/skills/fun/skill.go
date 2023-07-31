package fun

import (
	"embed"
	"io/fs"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
)

//go:embed assets
var fsAssets embed.FS

func getGenerators(generatorConfigs map[string]llm.GeneratorConfig) (generators llm.GeneratorMap, err error) {
	generators = llm.GeneratorMap{}
	for k, v := range generatorConfigs {
		switch v.Model() {
		case "gpt-3.5-turbo":
			generator, err := gpt.NewGenerator(gpt.WithConfig(v))
			if err != nil {
				return nil, err
			}
			generators[k] = generator
			continue
		}
		return nil, llm.ErrUnknownGeneratorModel
	}
	return
}

func New() (skill *gosk.Skill, err error) {
	subFS, err := fs.Sub(fsAssets, "assets")
	if err != nil {
		return
	}
	skill, err = gosk.ParseSemanticSkillFromFS(subFS, getGenerators)
	if err != nil {
		return
	}
	return
}
