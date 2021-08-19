package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	cmdTimeout = time.Second * 2
)

func main() {
	cmd := &cobra.Command{Use: "spire-pipe"}

	cmd.AddCommand(ConvertCommand())
	cmd.AddCommand(GenerateCommand())
	cmd.AddCommand(RPCCommand())
	cmd.AddCommand(DumpCommand())

	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintln(os.Stderr, "Did you forget to pipe in input?")
		}
		os.Exit(1)
	}
}
