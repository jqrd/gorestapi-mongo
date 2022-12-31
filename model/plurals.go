package model

import (
	"strings"

	pluralize "github.com/gertd/go-pluralize"
)

var pluralizeClient *pluralize.Client = pluralize.NewClient()

func EnsureCanonicalName(name string) string {
	return pluralizeClient.Singular(strings.TrimSpace(strings.ToLower(name)))
}
