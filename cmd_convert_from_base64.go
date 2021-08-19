package main

import (
	"context"
	"spire-pipe/codec"

	"github.com/spf13/cobra"
)

func ConvertFromBase64Command() *cobra.Command {
	impl := &convertFromBase64{
		inFormat: Base64Formats(),
	}
	cmd := &cobra.Command{
		Use:   "from-base64 TYPE",
		Short: "Converts from base64 to bytes",
		Args:  cobra.NoArgs,
		RunE:  runInOut(impl),
	}
	cmd.Flags().VarP(&impl.inFormat, "in-format", "", "input format")
	return cmd
}

type convertFromBase64 struct {
	inFormat BytesFormatFlag
}

func (cmd *convertFromBase64) Run(_ context.Context, in []byte, args []string) ([]byte, error) {
	return codec.BytesToBytes(in, cmd.inFormat, codec.RawBytes())
}
