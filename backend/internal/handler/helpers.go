package handler

import "encoding/json"

// marshalBody encodes the list payload for Redis caching.
// Extracted for testability and to keep the handler lean.
func marshalBody(body map[string]any) string {
	b, err := json.Marshal(body)
	if err != nil {
		return ""
	}
	return string(b)
}
