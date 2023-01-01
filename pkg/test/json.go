package test

import (
	"bytes"
	"encoding/json"
)

func JSON(obj interface{}) (string, error) {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(obj); err != nil {
		return "", err
	}

	return b.String(), nil
}
