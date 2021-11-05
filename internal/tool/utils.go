// Package tool is tomlMarshal
package tool

import (
	"bytes"

	bstoml "github.com/BurntSushi/toml"
)

func TomlMarshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bstoml.NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
