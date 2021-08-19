package main

import (
	"context"
	"spire-pipe/codec"

	"github.com/spf13/cobra"
)

func ConvertToPEMCommand() *cobra.Command {
	impl := &convertToPEM{
		inFormat: AllBytesFormats(),
	}
	cmd := &cobra.Command{
		Use:   "to-pem TYPE",
		Short: "Converts (optionally encoded) bytes to a PEM block",
		Args:  cobra.ExactArgs(1),
		RunE:  runInOut(impl),
	}
	cmd.Flags().VarP(&impl.inFormat, "in-format", "", "input format")
	return cmd
}

type convertToPEM struct {
	inFormat  BytesFormatFlag
	firstOnly bool
}

func (cmd *convertToPEM) Run(_ context.Context, in []byte, args []string) ([]byte, error) {
	pemType := args[0]
	return codec.BytesToBytes(in, cmd.inFormat, codec.PEMBytes(pemType, false))
}
