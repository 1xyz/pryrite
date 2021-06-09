package tools

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
)

// MarshalledError allows the error interface to serialized to json or yaml
// See: http://blog.magmalabs.io/2014/11/13/custom-error-marshaling-to-json-in-go.html for example
type MarshalledError struct {
	error
}

func NewMarshalledError(err error) *MarshalledError {
	return &MarshalledError{err}
}

func (e MarshalledError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Error())
}

// MarshalYAML marshal error to Yaml
// See: https://pkg.go.dev/gopkg.in/yaml.v2#Marshaler
func (e MarshalledError) MarshalYAML() (interface{}, error) {
	b, err := yaml.Marshal(e.Error())
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
