package llm

import (
	"encoding/json"
	"reflect"
	"strings"
)

func getNextPathPart(path string) (part string, remainingPath string) {
	idx := strings.Index(path, ".")
	if idx == -1 {
		return path, ""
	}
	part = path[0:idx]
	if idx < len(path)-1 {
		remainingPath = path[idx+1:]
	}
	return
}

func convertToMap(i interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var mapData map[string]interface{}
	err = json.Unmarshal(data, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

func setValue(m map[string]interface{}, key string, value interface{}) {
	if valueString, ok := value.(string); ok {
		var valueMap map[string]interface{}
		if err := json.Unmarshal([]byte(valueString), &valueMap); err == nil {
			m[key] = valueMap
			return
		}
		m[key] = valueString
		return
	}
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
		m[key] = value
		return
	}
	valueMap, err := convertToMap(value)
	if err == nil {
		m[key] = valueMap
		return
	}
	m[key] = value
}
