package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"net/url"

	"github.com/azdagron/spire-pipe/codec"
	"github.com/spf13/cobra"
)

func GenerateCSRCommand() *cobra.Command {
	impl := &generateCSR{
		keyFormat: AllBytesFormats(),
		csrFormat: AllBytesFormats(),
	}
	cmd := &cobra.Command{
		Use:   "csr",
		Short: "Generates a CSR from a key (provided on stdin)",
		Args:  cobra.ExactArgs(0),
		RunE:  runInOut(impl),
	}
	cmd.Flags().StringVarP(&impl.uriSAN, "uri-san", "", "", "URI SAN to include in the CSR")
	cmd.Flags().VarP(&impl.keyFormat, "key-format", "", "key input format")
	cmd.Flags().VarP(&impl.csrFormat, "csr-format", "", "CSR output format")
	return cmd
}

type generateCSR struct {
	uriSAN    string
	keyFormat BytesFormatFlag
	csrFormat BytesFormatFlag
}

func (cmd *generateCSR) Run(_ context.Context, in []byte, args []string) ([]byte, error) {
	uris := []*url.URL{}
	if cmd.uriSAN != "" {
		uri, err := url.Parse(cmd.uriSAN)
		if err != nil {
			return nil, fmt.Errorf("URI SAN is malformed: %v", err)
		}
		uris = append(uris, uri)
	}

	keyBytes, err := codec.BytesToBytes(in, cmd.keyFormat, codec.RawBytes())
	if err != nil {
		return nil, fmt.Errorf("key has invalid format: %v", err)
	}

	key, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("key is malformed: %v", err)
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		PublicKey: key.(crypto.Signer).Public(),
		URIs:      uris,
	}, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %v", err)
	}

	return codec.BytesToBytes(csrBytes, codec.RawBytes(), cmd.csrFormat)
}
