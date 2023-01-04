package conf

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/knadh/koanf/providers/env"
	"github.com/pkg/errors"
)

func GetEnvConfig[T interface{}](section string) (*T, error) {

	envProvider := env.Provider(section+".", ".", func(s string) string { return s })
	e, err := envProvider.Read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read env vars")
	}

	buffer, err := json.Marshal(e[section])
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not serialize %v.* env vars to JSON", section))
	}

	envConfig := new(T)
	err = json.Unmarshal(buffer, envConfig)
	if err != nil {
		t := reflect.TypeOf(*envConfig)
		return nil, errors.Wrap(err, fmt.Sprintf("could not deserialise %v.* env vars into %v", section, t.Name()))
	}

	return envConfig, nil

}
