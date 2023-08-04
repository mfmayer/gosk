package llm

import (
	"encoding/json"
	"fmt"
)

const (
	// TypeUser indicates that the message is user or human generated
	TypeUser string = "user"
	// TypeSystem indicates that the message is a system instruction to the model
	TypeSystem string = "system"
	// TypeNativeFunctionResponse indicates that the message contains a native function response
	TypeNativeFunctionResponse string = "functionResponse"
)

// Content can be
// * User Message -> Message.Role="user", Message.Content
// * System Message -> Message.Role="system", Message.Content
// * Assistant Message -> Message.Role="assistant", Message.Content
// * Function Call -> Message.Role="assistant", Message.FunctionCall.Name, Message.FunctionCall.Arguments
// * Function Response -> Message.Role="function", Massage.Content=<function response>
// * []Content with any above
type ContentRole string

const (
	RoleEmpty            ContentRole = ""
	RoleUser             ContentRole = "user"
	RoleSystem           ContentRole = "system"
	RoleAssistant        ContentRole = "assistant"
	RoleFunctionCall     ContentRole = "funcCall"
	RoleFunctionResponse ContentRole = "funcResponse"
)

type ContentProperty interface {
	fmt.Stringer
	// Set sets the property's value
	Set(value interface{}) Content
	// Value returns the property's value
	Value() interface{}
	// JSON returns the property's value as JSON
	JSON() []byte
}

type Content interface {
	ContentProperty
	// With sets a property at given path and is a shortcut for Property(path).Set(value)
	With(path string, value interface{}) Content
	// Prop returns the content's property at given path, nil if not available
	Property(path string) ContentProperty
	// Input is an alias for Content's Value()
	Input() ContentProperty
	// SetRole sets Content's optionl role
	SetRole(ContentRole) Content
	// Role returns Content's role, RoleEmpty if not set
	Role() ContentRole
	// SetName set Content's optional name
	SetName(string) Content
	// Name() return Content's name, "" if not set
	Name() string
	// WithPredecessor adds (or overwrites) `predecessor` option
	WithPredecessor(content Content) Content
	// Predecessor returns predecessor if "predecessor" option is available. Otherwise nil is returned
	Predecessor() Content
}

type content map[string]interface{}

// NewContent to create new Content with given content value
func NewContent(value ...interface{}) Content {
	content := content{}
	if len(value) <= 0 {
		return content
	}
	if len(value) == 1 {
		content.Set(value[0])
		return content
	}
	content.Set(value)
	return content
}

func (c content) value(path string) interface{} {
	if c == nil {
		return nil
	}
	if path == "" {
		// return root value
		v, ok := c[""]
		if !ok {
			return nil
		}
		return v
	}
	// return value at path
	currentContent := c
FINDVALUE:
	var current interface{} = map[string]interface{}(currentContent)
	for nextPathPart, remainingPath := getNextPathPart(path); nextPathPart != ""; nextPathPart, remainingPath = getNextPathPart(remainingPath) {
		// path not yet complete
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			// patch incomplete and current is not a map --> property not found in this content object
			current = nil
			break
		}
		// get next
		current, ok = currentMap[nextPathPart]
		if !ok {
			// patch incomplete and next element doesn't exist --> property not found in this content object
			current = nil
			break
		}
	}
	if current == nil {
		predecessor, ok := currentContent.Predecessor().(content)
		if ok {
			current = map[string]interface{}(predecessor)
			goto FINDVALUE
		}
	}
	// path complete
	return current
}

func (c content) string(path string) string {
	value := c.value(path)
	if value == nil {
		return ""
	}
	// return v
	if stringValue, ok := value.(string); ok {
		return stringValue
	}
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(marshalledValue)
}

func (c content) String() string {
	return c.string("")
}

func (c content) Set(value interface{}) Content {
	return c.With("", value)
}

func (c content) Value() interface{} {
	return c.value("")
}

func (c content) With(path string, value interface{}) Content {
	if c == nil {
		return nil
	}
	if value == nil {
		// nothing to set
		return nil
	}
	if path == "" {
		// set value at root
		setValue(c, "", value)
		return c
	}
	// set value at path
	var currentMap map[string]interface{} = c
	for nextPathPart, remainingPath := getNextPathPart(path); nextPathPart != "" || remainingPath != ""; nextPathPart, remainingPath = getNextPathPart(remainingPath) {
		// Last path part? --> Set value
		if len(remainingPath) <= 0 {
			setValue(currentMap, nextPathPart, value)
			break
		}
		// get next
		next, ok := currentMap[nextPathPart]
		if !ok {
			// next element doesn't exist --> create map and set it
			nextMap := map[string]interface{}{}
			currentMap[nextPathPart] = nextMap
			currentMap = nextMap
			continue
		}
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			// next element is not a map --> create map and replace it
			nextMap := map[string]interface{}{}
			currentMap[nextPathPart] = nextMap
			currentMap = nextMap
			continue
		}
		// next element is already a map, set as current map
		currentMap = nextMap
	}
	// path complete
	return c
}

func (c content) Property(path string) ContentProperty {
	return contentEntry{
		path: path,
		cm:   c,
	}
}

func (c content) Input() ContentProperty {
	return c.Property("")
}

func (c content) SetRole(role ContentRole) Content {
	c.With("role", role)
	return c
}

func (c content) Role() ContentRole {
	if role, ok := c.value("role").(ContentRole); ok {
		return role
	}
	return ""
}

func (c content) SetName(name string) Content {
	c.With("name", name)
	return c
}

func (c content) Name() string {
	if name, ok := c.value("name").(string); ok {
		return name
	}
	return ""
}

func (c content) JSON() []byte {
	if c == nil {
		return nil
	}
	marshalledValue, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	return marshalledValue
}

func (c content) WithPredecessor(content Content) Content {
	c["predecessor"] = content
	return c
}

func (c content) Predecessor() Content {
	if predecessor, ok := c.value("predecessor").(Content); ok {
		return predecessor
	}
	return nil
}

type contentEntry struct {
	path string
	cm   content
}

func (ce contentEntry) String() string {
	if ce.cm == nil {
		return ""
	}
	return ce.cm.string(ce.path)
}

func (ce contentEntry) JSON() []byte {
	if ce.cm == nil {
		return nil
	}
	value := ce.cm.value(ce.path)
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return bytes
}

func (ce contentEntry) Value() interface{} {
	if ce.cm == nil {
		return nil
	}
	return ce.cm.value(ce.path)
}

func (ce contentEntry) Set(value interface{}) Content {
	if ce.cm == nil {
		return nil
	}
	ce.cm.With(ce.path, value)
	return ce.cm
}
