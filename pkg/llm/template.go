package llm

import (
	"bytes"
	"fmt"
	"io/fs"
	"text/template"
)

func TemplateFromFS(fsys fs.FS, patterns ...string) (*template.Template, error) {
	template, err := template.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, err
	}
	promptTemplate := template.Lookup("skprompt.tmpl")
	if promptTemplate != nil {
		return promptTemplate, nil
	}
	return template, fmt.Errorf("\"skprompt.tmpl\" not found")
}

func TemplateFromText(text string) (*template.Template, error) {
	return template.New("skprompt").Parse(text)
}

func ExecuteTemplate(template *template.Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := template.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), err
}

func ApplyTemplateToContent(template *template.Template, content Content) error {
	text, err := ExecuteTemplate(template, content)
	if err != nil {
		content.Set(text)
	}
	return err
}
