package llm

// Generator as a generic interface for large langage model generators
type Generator interface {
	GenerateResponse(input Content) (response Content, err error)
}

type GeneratorMap map[string]Generator

// FindAny returns first of given comma separated generators in map
func (m GeneratorMap) FindAny(generators ...string) (Generator, bool) {
	// tokenList := strings.Split(generatorsCommaSeperated, ",")
	for _, generator := range generators {
		// token = strings.TrimSpace(token) // Removes any leading/trailing spaces from token
		if val, ok := m[generator]; ok {
			return val, true
		}
	}
	return nil, false
}
