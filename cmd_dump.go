package main

import "github.com/spf13/cobra"

func DumpCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "dump"}
	cmd.AddCommand(DumpX509SVIDIDCommand())
	cmd.AddCommand(DumpJWTSVIDIDCommand())
	return cmd
}
