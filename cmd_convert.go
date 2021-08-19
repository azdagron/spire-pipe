package main

import "github.com/spf13/cobra"

func ConvertCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "convert", Aliases: []string{"conv"}}
	cmd.AddCommand(ConvertToPEMCommand())
	cmd.AddCommand(ConvertFromPEMCommand())
	cmd.AddCommand(ConvertToBase64Command())
	cmd.AddCommand(ConvertFromBase64Command())
	return cmd
}
