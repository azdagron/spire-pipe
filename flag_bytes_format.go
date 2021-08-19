package main

import (
	"errors"
	"fmt"
	"spire-pipe/codec"
	"strings"
)

type BytesFormatFlag []codec.Bytes

func AllBytesFormats() []codec.Bytes {
	return append(Base64Formats(), codec.RawBytes())
}

func Base64Formats() []codec.Bytes {
	return []codec.Bytes{
		codec.StdBase64Bytes(),
		codec.URLBase64Bytes(),
		codec.RawStdBase64Bytes(),
		codec.RawURLBase64Bytes(),
	}
}

func (fl BytesFormatFlag) Set(value string) error {
	if len(fl) == 0 {
		return errors.New("flag misconfigured internally")
	}
	name := strings.ToLower(value)
	for i, choice := range fl {
		if choice.Name() == name {
			fl[i] = fl[0]
			fl[0] = choice
			return nil
		}
	}
	return fmt.Errorf("unknown format %q", value)
}

func (fl BytesFormatFlag) Type() string {
	return "format"
}

func (fl BytesFormatFlag) String() string {
	return fl.selection().Name()
}

func (fl BytesFormatFlag) Name() string {
	return fl.selection().Name()
}

func (fl BytesFormatFlag) BytesIn(b []byte) ([]byte, error) {
	return fl.selection().BytesIn(b)
}

func (fl BytesFormatFlag) BytesOut(b []byte) ([]byte, error) {
	return fl.selection().BytesOut(b)
}

func (fl BytesFormatFlag) selection() codec.Bytes {
	if len(fl) == 0 {
		return unsetBytes{}
	}
	return fl[0]
}

type unsetBytes struct{}

func (unsetBytes) Name() string { return "" }
func (unsetBytes) BytesIn([]byte) ([]byte, error) {
	return nil, errors.New("internal: flag is not initialized")
}
func (unsetBytes) BytesOut([]byte) ([]byte, error) {
	return nil, errors.New("internal: flag is not initialized")
}
