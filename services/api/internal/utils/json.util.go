package utils

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

func ParseJSON[T any](raw []byte) (T, error) {
	var result T
	if len(raw) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, err
	}
	return result, nil
}

// Structured Outputs uses a subset of JSON schema
// These flags are necessary to comply with the subset
func GenerateSchema[T any]() map[string]any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)

	data, _ := json.Marshal(schema)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}
