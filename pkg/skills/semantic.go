package skills

import (
	"embed"
	"io/fs"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/llm"
)

//go:embed assets/*
var semanticSkillsDir embed.FS

func CreateSemanticSkills(generators map[string]llm.Generator, skillNames ...string) (skills map[string]*gosk.Skill, err error) {
	assetsDir, err := fs.Sub(semanticSkillsDir, "assets")
	if err != nil {
		return
	}
	return gosk.ParseSemanticSkillsFromFS(assetsDir, generators, skillNames...)
}
