package llm

import (
	"testing"
)

func TestFindEntry(t *testing.T) {

	// m := map[string]interface{}{
	// 	"foo": "bar",
	// 	"barm": map[string]interface{}{
	// 		"fooe": "bare",
	// 	},
	// }
	// e := findEntry(m, "fooe")
	// t.Log(e)
	s := struct {
		Test int
		Bla  string
	}{10, "blubb"}

	c := NewContent("hallo").SetRole(RoleUser).SetName("Hans").Set("prompt text").With("foo.bar", 5).With("foo.faa.fee", "{\"testVal\":5}").With("foo.faa", s)
	entry := c.Property("foo.faa.Test")
	entryValue := entry.Value()
	t.Logf("Content: %v", entryValue)
	t.Logf("Content: %s", c)

}
