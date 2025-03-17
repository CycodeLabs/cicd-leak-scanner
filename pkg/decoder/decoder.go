package decoder

import (
	"fmt"

	"github.com/CycodeLabs/cicd-leak-scanner/pkg/decoder/base64"
)

type Decoder interface {
	Id() string
	Decode(input string, repeat int) (string, error)
}

func New(id string) (Decoder, error) {
	switch id {
	case base64.Id:
		return base64.New()
	default:
		return nil, fmt.Errorf("unknown analyzer: %s", id)
	}
}
