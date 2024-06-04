package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/azdagron/spire-pipe/codec"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/spf13/cobra"
)

func DumpJWTSVIDIDCommand() *cobra.Command {
	impl := &dumpJWTSVIDID{
		svidFormat: AllBytesFormats(),
	}
	cmd := &cobra.Command{
		Use:   "jwt-svid-id",
		Short: "Dumps the SPIFFE ID of an JWT-SVID",
		Args:  cobra.NoArgs,
		RunE:  runInOut(impl),
	}
	cmd.Flags().VarP(&impl.svidFormat, "svid-format", "", "JWT-SVID format")
	return cmd
}

type dumpJWTSVIDID struct {
	svidFormat BytesFormatFlag
}

func (cmd *dumpJWTSVIDID) Run(_ context.Context, in []byte, args []string) ([]byte, error) {
	svidBytes, err := codec.BytesToBytes(in, cmd.svidFormat, codec.RawBytes())
	if err != nil {
		return nil, fmt.Errorf("JWT-SVID has invalid format: %v", err)
	}

	tok, err := jwt.ParseSigned(string(svidBytes), []jose.SignatureAlgorithm{
		jose.RS256, jose.RS384, jose.RS512,
		jose.ES256, jose.ES384, jose.ES512,
		jose.PS256, jose.PS384, jose.PS512,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse JWT-SIVD: %v", err)
	}

	var claims jwt.Claims
	if err := tok.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, fmt.Errorf("unable to get claims from JWT-SVID: %v", err)
	}

	if len(claims.Subject) == 0 {
		return nil, errors.New("JWT-SVID missing SPIFFE ID claim")
	}

	return []byte(claims.Subject + "\n"), nil
}
