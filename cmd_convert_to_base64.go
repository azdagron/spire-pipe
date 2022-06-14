package main

import (
	"context"

	"github.com/azdagron/spire-pipe/codec"
	"github.com/spf13/cobra"
)

func ConvertToBase64Command() *cobra.Command {
	impl := &convertToBase64{
		outFormat: Base64Formats(),
	}
	cmd := &cobra.Command{
		Use:   "to-base64 TYPE",
		Short: "Converts from bytes to base64",
		Args:  cobra.NoArgs,
		RunE:  runInOut(impl),
	}
	cmd.Flags().VarP(&impl.outFormat, "out-format", "", "output format")
	return cmd
}

type convertToBase64 struct {
	outFormat BytesFormatFlag
}

func (cmd *convertToBase64) Run(_ context.Context, in []byte, args []string) ([]byte, error) {
	return codec.BytesToBytes(in, codec.RawBytes(), cmd.outFormat)
}
