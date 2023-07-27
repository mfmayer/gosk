package utils

import (
	"errors"
	"strings"
)

// GetValue retrieves a value from a nested map given a dot-separated path (e.g. "a.b.c").
// It returns the retrieved value and a boolean indicating the success of the operation.
func GetValue(m map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var ok bool
	var val interface{} = m
	for _, key := range keys {
		m, ok = val.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	return val, true
}

// setValue sets a value in a nested map given a dot-separated path. If the path does not exist, it is created.
// If any part of the path exists and is not a map, an error is returned.
// This function creates any necessary maps if they don't exist, but does not overwrite non-map values in the path.
func SetValue(m map[string]interface{}, path string, value interface{}) error {
	keys := strings.Split(path, ".")

	for i := range keys {
		if i == len(keys)-1 {
			m[keys[i]] = value
			return nil
		} else {
			// If the key doesn't exist or isn't a map, create a new map.
			if _, ok := m[keys[i]].(map[string]interface{}); !ok {
				// But if the key already exists with a non-map value, return an error.
				if _, exists := m[keys[i]]; exists {
					return errors.New("non-map key already exists in path")
				}
				m[keys[i]] = make(map[string]interface{})
			}
		}
		m = m[keys[i]].(map[string]interface{})
	}
	return nil
}
