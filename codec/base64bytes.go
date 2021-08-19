package codec

import "encoding/base64"

func StdBase64Bytes() Bytes {
	return base64Bytes{encoding: base64.StdEncoding, name: "std-base64"}
}

func RawStdBase64Bytes() Bytes {
	return base64Bytes{encoding: base64.RawStdEncoding, name: "raw-std-base64"}
}

func URLBase64Bytes() Bytes {
	return base64Bytes{encoding: base64.URLEncoding, name: "url-base64"}
}

func RawURLBase64Bytes() Bytes {
	return base64Bytes{encoding: base64.RawURLEncoding, name: "raw-url-base64"}
}

type base64Bytes struct {
	name     string
	encoding *base64.Encoding
}

func (c base64Bytes) Name() string { return c.name }

func (c base64Bytes) BytesIn(in []byte) ([]byte, error) {
	out := make([]byte, c.encoding.DecodedLen(len(in)))
	n, err := c.encoding.Decode(out, in)
	if err != nil {
		return nil, err
	}
	return out[:n], nil
}

func (c base64Bytes) BytesOut(in []byte) ([]byte, error) {
	out := make([]byte, c.encoding.EncodedLen(len(in)))
	c.encoding.Encode(out, in)
	return out, nil
}
