package codec

func RawBytes() Bytes {
	return rawBytes{}
}

type rawBytes struct{}

func (rawBytes) Name() string                       { return "raw" }
func (rawBytes) BytesIn(in []byte) ([]byte, error)  { return in, nil }
func (rawBytes) BytesOut(in []byte) ([]byte, error) { return in, nil }
