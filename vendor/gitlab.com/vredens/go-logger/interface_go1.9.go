// +build go1.9

package logger

// Fields is a type alias for map[string]interface{}.
// @deprecated use KV instead.
type Fields = map[string]interface{}

// KV is a type alias for map[string]interface{}.
type KV = map[string]interface{}

// Tags is a type alias for []string
type Tags = []string
