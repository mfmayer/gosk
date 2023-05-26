package gosk

type SkillFunction func(parameters ...string) (string, error)

type Skill map[string]SkillFunction
