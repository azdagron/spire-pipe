package codec

type Bytes interface {
	Name() string
	BytesIn([]byte) ([]byte, error)
	BytesOut([]byte) ([]byte, error)
}

func BytesToBytes(in []byte, inCodec, outCodec Bytes) ([]byte, error) {
	data, err := inCodec.BytesIn(in)
	if err != nil {
		return nil, err
	}
	return outCodec.BytesOut(data)
}
