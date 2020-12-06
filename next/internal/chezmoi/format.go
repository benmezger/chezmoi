package chezmoi

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// A Format is a serialization format.
type Format interface {
	Decode(data []byte, value interface{}) error
	Marshal(value interface{}) ([]byte, error)
	Unmarshal(data []byte) (interface{}, error)
}

// A jsonFormat implements the JSON serialization format.
type jsonFormat struct{}

// A tomlFormat implements the TOML serialization format.
type tomlFormat struct{}

// A yamlFormat implements the YAML serialization format.
type yamlFormat struct{}

// Formats is a map of all Formats by name.
var Formats = map[string]Format{
	"json": jsonFormat{},
	"toml": tomlFormat{},
	"yaml": yamlFormat{},
}

// Decode implements Format.Decode.
func (jsonFormat) Decode(data []byte, value interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(data)).Decode(value)
}

// Marshal implements Format.Marshal.
func (jsonFormat) Marshal(value interface{}) ([]byte, error) {
	sb := strings.Builder{}
	e := json.NewEncoder(&sb)
	e.SetIndent("", "  ")
	if err := e.Encode(value); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

// Unmarshal implements Format.Unmarshal.
func (jsonFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Decode implements Format.Decode.
func (tomlFormat) Decode(data []byte, value interface{}) error {
	return toml.NewDecoder(bytes.NewBuffer(data)).Decode(value)
}

// Marshal implements Format.Marshal.
func (tomlFormat) Marshal(value interface{}) ([]byte, error) {
	return toml.Marshal(value)
}

// Unmarshal implements Format.Unmarshal.
func (tomlFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Decode implements Format.Decode.
func (yamlFormat) Decode(data []byte, value interface{}) error {
	return yaml.NewDecoder(bytes.NewBuffer(data)).Decode(value)
}

// Marshal implements Format.Marshal.
func (yamlFormat) Marshal(value interface{}) ([]byte, error) {
	return yaml.Marshal(value)
}

// Unmarshal implements Format.Unmarshal.
func (yamlFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
