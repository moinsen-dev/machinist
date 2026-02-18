package domain

import (
	"bytes"
	"os"

	"github.com/BurntSushi/toml"
)

// MarshalManifest serializes a Snapshot to TOML bytes.
func MarshalManifest(s *Snapshot) ([]byte, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalManifest deserializes TOML bytes into a Snapshot.
func UnmarshalManifest(data []byte) (*Snapshot, error) {
	var s Snapshot
	if _, err := toml.Decode(string(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WriteManifest writes a Snapshot as TOML to the given file path.
func WriteManifest(s *Snapshot, path string) error {
	data, err := MarshalManifest(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ReadManifest reads a TOML manifest file and returns a Snapshot.
func ReadManifest(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return UnmarshalManifest(data)
}
