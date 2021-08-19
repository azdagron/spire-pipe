package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"spire-pipe/codec"

	"github.com/spf13/cobra"
)

func DumpX509SVIDIDCommand() *cobra.Command {
	impl := &dumpX509SVIDID{
		svidFormat: AllBytesFormats(),
	}
	cmd := &cobra.Command{
		Use:   "x509-svid-id",
		Short: "Dumps the SPIFFE ID of an X509-SVID",
		Args:  cobra.NoArgs,
		RunE:  runInOut(impl),
	}
	cmd.Flags().VarP(&impl.svidFormat, "svid-format", "", "X509-SVID format")
	return cmd
}

type dumpX509SVIDID struct {
	svidFormat BytesFormatFlag
}

func (cmd *dumpX509SVIDID) Run(_ context.Context, in []byte, args []string) ([]byte, error) {
	svidBytes, err := codec.BytesToBytes(in, cmd.svidFormat, codec.RawBytes())
	if err != nil {
		return nil, fmt.Errorf("X509-SVID has invalid format: %v", err)
	}

	certs, err := x509.ParseCertificates(svidBytes)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 {
		return nil, errors.New("empty input")
	}

	cert := certs[0]
	if len(cert.URIs) != 1 {
		return nil, fmt.Errorf("expected one URI SAN; got %d", len(cert.URIs))
	}

	return []byte(cert.URIs[0].String() + "\n"), nil
}
