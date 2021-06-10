package tools

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"time"
)

// MarshalledError allows the error interface to serialized to json or yaml
// See: http://blog.magmalabs.io/2014/11/13/custom-error-marshaling-to-json-in-go.html for example
// ToDo: UnMarshall Yaml & JSON
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

// MarshalledDuration allows the duration to de-serialized from Yaml
// ToDo: json and Marshall Yaml
type MarshalledDuration time.Duration

// UnmarshalYAML unmarshalls Yaml to duration
// See: https://pkg.go.dev/gopkg.in/yaml.v2#Unmarshaler
func (t *MarshalledDuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tm string
	if err := unmarshal(&tm); err != nil {
		return err
	}

	td, err := time.ParseDuration(tm)
	if err != nil {
		return fmt.Errorf("failed to parse '%s' to time.Duration: %v", tm, err)
	}

	*t = MarshalledDuration(td)
	return nil
}

func (t *MarshalledDuration) Duration() time.Duration {
	return time.Duration(*t)
}
