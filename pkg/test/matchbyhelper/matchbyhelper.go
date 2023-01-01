package matchbyhelper

// workaround for https://github.com/stretchr/testify/issues/504
// see example usage in thing_test.go

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MatchHelper struct {
	isAsserting bool
}

func New() *MatchHelper {
	return &MatchHelper{}
}

func (m *MatchHelper) BeginAssert() {
	m.isAsserting = true
}

func MockMatchedBy[T interface{}](t *testing.T, m *MatchHelper, matcher func(actual T) bool) interface{} {
	calls := make([]string, 0)
	replayCallIndex := 0
	return mock.MatchedBy(func(actual T) bool {
		jsonText, err := valueToJson(actual)
		if err != nil {
			t.Errorf("MockMatchedBy: could not convert actual value to JSON. Error: %v.", err)
			return false
		}

		if !m.isAsserting {
			calls = append(calls, jsonText)
		} else {
			actual, err = valueFromJson[T](calls[replayCallIndex])
			replayCallIndex++

			if err != nil {
				t.Errorf("MockMatchedBy: could not convert JSON argument from previous call during replay. Error: %v.", err)
				return false
			}
		}

		return matcher(actual)
	})
}

func valueToJson(obj interface{}) (string, error) {
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(obj); err != nil {
		return "", err
	}

	return b.String(), nil
}

func valueFromJson[T interface{}](jsonText string) (T, error) {
	result := new(T)
	err := json.NewDecoder(strings.NewReader(jsonText)).Decode(result)
	return *result, err
}
