package main

import "github.com/spf13/cobra"

func GenerateCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "generate", Aliases: []string{"gen"}}
	cmd.AddCommand(GenerateKeyCommand())
	cmd.AddCommand(GenerateCSRCommand())
	return cmd
}
