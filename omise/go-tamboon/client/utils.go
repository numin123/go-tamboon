package client

import (
	"encoding/json"
)

func parseOmiseError(body []byte) string {
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return "unknown error"
	}

	if object, ok := errorResponse["object"].(string); ok && object == "error" {
		if message, ok := errorResponse["message"].(string); ok {
			return message
		}
	}

	return "unknown error"
}
