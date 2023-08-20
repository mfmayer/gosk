package test

import (
	"testing"

	"github.com/mfmayer/gosk/pkg/llm"
)

func TestFindEntry(t *testing.T) {
	testStruct := struct {
		Test int
		Bla  string
	}{10, "blubb"}

	c := llm.NewContent("hallo").
		SetRole(llm.RoleUser).
		SetName("Hans").
		Set("prompt text").
		With("foo.bar", 5).
		With("foo.faa.fee", "{\"testVal\":5}").
		With("foo.faa", testStruct)
	entry := c.Property("foo.faa.Test")
	entryValue := entry.Value()
	t.Logf("Content: %v", entryValue)
	t.Logf("Content: %s", c)
}

func TestPredecessor(t *testing.T) {
	c1 := llm.NewContent().With("bar", "foo")
	c2 := llm.NewContent().With("foo", "bar").WithPredecessor(c1)
	bar := c2.Property("foo")
	t.Log(bar.Value())
}
