package main

import (
	"context"
	"spire-pipe/codec"

	"github.com/spf13/cobra"
)

func ConvertFromPEMCommand() *cobra.Command {
	impl := &convertFromPEM{
		outFormat: AllBytesFormats(),
	}
	cmd := &cobra.Command{
		Use:   "from-pem TYPE",
		Short: "Converts a PEM block to (optionally encoded) bytes",
		Args:  cobra.ExactArgs(1),
		RunE:  runInOut(impl),
	}
	cmd.Flags().VarP(&impl.outFormat, "out-format", "", "output format")
	cmd.Flags().BoolVarP(&impl.firstOnly, "first-only", "", false, "first block only")
	return cmd
}

type convertFromPEM struct {
	outFormat BytesFormatFlag
	firstOnly bool
}

func (cmd *convertFromPEM) Run(_ context.Context, in []byte, args []string) ([]byte, error) {
	pemType := args[0]
	return codec.BytesToBytes(in, codec.PEMBytes(pemType, cmd.firstOnly), cmd.outFormat)
}
