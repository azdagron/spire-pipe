package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"spire-pipe/codec"

	"github.com/spf13/cobra"
)

func GenerateKeyCommand() *cobra.Command {
	impl := &generateKey{
		outFormat: AllBytesFormats(),
	}
	cmd := &cobra.Command{
		Use:  "key",
		Args: cobra.ExactArgs(0),
		RunE: runOut(impl),
	}
	cmd.Flags().VarP(&impl.outFormat, "out-format", "", "output format")
	return cmd
}

type generateKey struct {
	outFormat BytesFormatFlag
}

func (cmd *generateKey) Run(_ context.Context, args []string) ([]byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	return codec.BytesToBytes(keyBytes, codec.RawBytes(), cmd.outFormat)
}
