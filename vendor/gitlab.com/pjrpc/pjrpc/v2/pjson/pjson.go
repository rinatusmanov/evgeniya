//go:build !js && !wasm

// Package pjson is currently using the std golang package.
// It will be replaced in the future.
package pjson

import (
	"encoding/json"
	"io"
)

// Marshal returns the JSON encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v) // nolint:wrapcheck // No additional error context.
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v) // nolint:wrapcheck // No additional error context.
}

// Decoder is a wrapper if the JSON decoder.
type Decoder struct {
	Dec *json.Decoder
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{Dec: json.NewDecoder(r)}
}

// Decode reads the next JSON-encoded value from its
// input and stores it in the value pointed to by v.
func (d *Decoder) Decode(v interface{}) error {
	return d.Dec.Decode(v) // nolint:wrapcheck // No additional error context.
}

// Encoder wrapper of the JSON encoder.
type Encoder struct {
	Enc *json.Encoder
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{Enc: json.NewEncoder(w)}
}

// Encode writes the JSON encoding of v to the stream,
// followed by a newline character.
func (e *Encoder) Encode(v interface{}) error {
	return e.Enc.Encode(v) // nolint:wrapcheck // No additional error context.
}
