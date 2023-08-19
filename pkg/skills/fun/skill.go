package fun

import (
	"embed"
	"io/fs"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/llm"
)

//go:embed assets
var fsAssets embed.FS

func New(generatorFactories llm.GeneratorFactoryMap) (skill *gosk.Skill, err error) {
	subFS, err := fs.Sub(fsAssets, "assets")
	if err != nil {
		return
	}
	skill, err = gosk.ParseSemanticSkillFromFS(subFS, generatorFactories)
	if err != nil {
		return
	}
	return
}
