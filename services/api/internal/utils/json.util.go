package utils

import "encoding/json"

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
