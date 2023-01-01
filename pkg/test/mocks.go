package test

import (
	"testing"
)

func MatchJson[T interface{}](t *testing.T, expected T) func(actual T) bool {
	return func(actual T) bool {
		expectedJson, expectedErr := JSON(expected)
		actualJson, actualErr := JSON(actual)
		if expectedErr != nil || actualErr != nil {
			if expectedErr != nil {
				t.Errorf("mock.MatchedBy: could not convert expected value to JSON. Error: %v.", expectedErr)
			}
			if actualErr != nil {
				t.Errorf("mock.MatchedBy: could not convert actual value to JSON. Error: %v.", actualErr)
			}
			return false
		}

		if expectedJson != actualJson {
			t.Errorf("mock.MatchedBy not matched. Expected: %v. Actual: %v.", expectedJson, actualJson)
			return false
		}

		return true
	}
}
