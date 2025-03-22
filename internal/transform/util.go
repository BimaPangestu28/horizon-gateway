package transform

import (
	"encoding/json"
)

// isJSON checks if data is valid JSON
func isJSON(data []byte) bool {
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}
