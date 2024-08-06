package main

import (
	"context"
	"io"

	"github.com/spf13/cobra"
)

type outCommand interface {
	Run(ctx context.Context, args []string) ([]byte, error)
}

type inOutCommand interface {
	Run(ctx context.Context, in []byte, args []string) ([]byte, error)
}

func runOut(cmd outCommand) func(cobraCmd *cobra.Command, args []string) error {
	return func(cobraCmd *cobra.Command, args []string) error {
		cobraCmd.SilenceUsage = true

		out, err := cmd.Run(cobraCmd.Context(), args)
		if err != nil {
			return err
		}
		if _, err := cobraCmd.OutOrStdout().Write(out); err != nil {
			return err
		}
		return nil
	}
}

func runInOut(cmd inOutCommand) func(cobraCmd *cobra.Command, args []string) error {
	return func(cobraCmd *cobra.Command, args []string) error {
		cobraCmd.SilenceUsage = true

		in, err := readAll(cobraCmd.Context(), cobraCmd.InOrStdin())
		if err != nil {
			return err
		}
		out, err := cmd.Run(cobraCmd.Context(), in, args)
		if err != nil {
			return err
		}
		if _, err := cobraCmd.OutOrStdout().Write(out); err != nil {
			return err
		}
		return nil
	}
}

func readAll(ctx context.Context, r io.Reader) ([]byte, error) {
	type result struct {
		in  []byte
		err error
	}
	resultCh := make(chan result, 1)
	go func() {
		in, err := io.ReadAll(r)
		resultCh <- result{in: in, err: err}
	}()

	select {
	case r := <-resultCh:
		return r.in, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
