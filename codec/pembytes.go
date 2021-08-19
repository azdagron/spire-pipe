package codec

import (
	"encoding/pem"
	"errors"
	"fmt"
)

func PEMBytes(pemType string, firstOnly bool) Bytes {
	return pemBytes{pemType: pemType, firstOnly: firstOnly}
}

type pemBytes struct {
	pemType   string
	firstOnly bool
}

func (pemBytes) Name() string { return "pem" }

func (c pemBytes) BytesIn(in []byte) ([]byte, error) {
	block, rest := pem.Decode(in)
	if block == nil {
		return nil, errors.New("input is not PEM")
	}

	nextBlock, _ := pem.Decode(rest)
	if nextBlock != nil && !c.firstOnly {
		return nil, errors.New("only one PEM block expected on input")
	}

	if block.Type != string(c.pemType) {
		return nil, fmt.Errorf("expected %q PEM block; got %q", c.pemType, block.Type)
	}

	return block.Bytes, nil
}

func (c pemBytes) BytesOut(in []byte) ([]byte, error) {
	return pem.EncodeToMemory(&pem.Block{
		Type:  string(c.pemType),
		Bytes: in,
	}), nil
}
