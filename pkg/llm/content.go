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
	//TODO: Define how content looks like
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

type Content interface {
	fmt.Stringer
	// Data returns Content's "data" object which represents the content's payload.
	Data() interface{} //map[string]interface{}
	// StringData returns Content's payload in a marshalled (serialized) json format when possible.
	// If content is a string or a fmt.Stringer, the according string is directly returned.
	StringData() string
	// WithData adds (or overwrites) the content's data property.
	WithData(data interface{}) Content
	// With adds (or overwrites) option data for given name. Attention using "data" as name will overwrite the content's data payload.
	With(name string, optionData interface{}) Content
	// Option returns any set option and true if found in this or any predecessors. If not found nil is returned.
	Option(name string) interface{}
	// WithRoleOption adds (or overwrites) role option
	WithRoleOption(role ContentRole) Content
	// RoleOption returns ContentRole if "role" option is available. Otherwise RoleEmpty (empty string) is returned.
	RoleOption() ContentRole
	// WithNameOption adds (or overwrites) name option
	WithNameOption(name string) Content
	// NameOption resturns name if "name" option is available. Otherwise empty string is returned.
	NameOption() string
	// WithPredecessor adds (or overwrites) `predecessor` option
	WithPredecessor(content Content) Content
	// Predecessor returns predecessor if "predecessor" option is available. Otherwise nil is returned
	Predecessor() Content
	// IsStructured returns true if data is structured
	// IsStructured() bool
}

type content map[string]interface{}

func (c content) String() string {
	return c.StringData()
}

func (c content) Data() interface{} {
	if c == nil {
		return nil
	}
	return c.Option("data")
}

func (c content) StringData() string {
	data := c.Data()
	if data == nil {
		return ""
	}
	switch d := data.(type) {
	case string:
		return d
	case fmt.Stringer:
		return d.String()
	}
	if jsonData, err := json.Marshal(data); err == nil {
		return string(jsonData)
	}
	return ""
}

func (c content) WithData(data interface{}) Content {
	if c == nil {
		return c
	}
	if data == nil {
		return c.With("data", nil)
	}
	if d, ok := data.(string); ok {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(d), &m); err == nil {
			return c.With("data", m)
		}
	}
	return c.With("data", data)
}

func (c content) With(name string, optionData interface{}) Content {
	if c == nil {
		return c
	}
	c[name] = optionData
	return c
}

func (c content) Option(name string) interface{} {
	if c == nil {
		return nil
	}
	value, ok := c[name]
	if !ok && name != "predecessor" {
		if predecessor := c.Predecessor(); predecessor != nil {
			return predecessor.Option(name)
		}
	}
	return value
}

func (c content) WithRoleOption(role ContentRole) Content {
	return c.With("role", role)
}

func (c content) RoleOption() ContentRole {
	role := c.Option("role")
	if role == nil {
		return RoleEmpty
	}
	if typedRole, ok := role.(ContentRole); ok {
		return typedRole
	}
	return RoleEmpty
}

func (c content) WithNameOption(name string) Content {
	return c.With("name", name)
}

func (c content) NameOption() string {
	name := c.Option("name")
	if name == nil {
		return ""
	}
	if typedName, ok := name.(string); ok {
		return typedName
	}
	return ""
}

func (c content) WithPredecessor(content Content) Content {
	return c.With("predecessor", content)
}

func (c content) Predecessor() Content {
	predecessor := c.Option("predecessor")
	if predecessor == nil {
		return nil
	}
	if typedPredecessor, ok := predecessor.(Content); ok {
		return typedPredecessor
	}
	return nil
}

// NewContent to create new Content. As data only string, fmt.Stringer or map[string]interface{} structures are supported
// JSON string representations will be automatically structurized so that IsStructure() would be also true
func NewContent(data ...interface{}) Content {
	content := content{}
	if len(data) == 0 {
		return content
	}
	if len(data) == 1 {
		return content.WithData(data[0])
	}
	return content.WithData(data)
}
