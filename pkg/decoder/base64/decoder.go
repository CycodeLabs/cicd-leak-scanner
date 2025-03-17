package base64

import (
	"encoding/base64"
	"fmt"
)

const (
	Id = "base64_decode"
)

type Base64 struct{}

func New() (*Base64, error) {
	return &Base64{}, nil
}

func (b *Base64) Decode(input string, repeat int) (string, error) {
	for i := 0; i < repeat; i++ {
		decoded, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}
		input = string(decoded)
	}

	return input, nil
}

func (b *Base64) Id() string {
	return "base64"
}
