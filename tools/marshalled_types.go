package tools

import (
	"fmt"
	"time"
)

// MarshalledDuration allows the duration to de-serialized from Yaml
// ToDo: json and Marshall Yaml
type MarshalledDuration struct {
	time.Duration
}

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

	t.Duration = td
	return nil
}

func (t MarshalledDuration) MarshalYAML() (interface{}, error) {
	return t.Duration.String(), nil
}

func (t MarshalledDuration) GetDuration() time.Duration {
	return t.Duration
}
