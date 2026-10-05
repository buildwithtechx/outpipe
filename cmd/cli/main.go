package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
)

var version = "dev"

var errUnknownCommand = fmt.Errorf("unknown command")

func main() {

	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "outpipe: %v\n", err)

		if errors.Is(err, errUnknownCommand) {
			os.Exit(2)
		}

		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := loadCLISettings()

	if err != nil {
		return err
	}

	root := newRootCommand(cfg)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {

		if strings.Contains(err.Error(), "unknown command") {
			return fmt.Errorf("%w %q", errUnknownCommand, args[0])
		}

		return err
	}

	return nil
}

func newRootCommand(cfg config.CLIConfig) *cobra.Command {
	root := &cobra.Command{
		Use:           "outpipe",
		Short:         "outpipe tunnels connections from your machine to the public network",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newVersionCommand(),
		newHealthCommand(cfg),
		newLoginCommand(cfg),
		newOpenCommand(cfg),
		newTCPCommand(cfg),
		newTunnelCommand(cfg, "create", "create a managed tunnel", cobra.ExactArgs(1)),
		newTunnelCommand(cfg, "list", "list managed tunnels", cobra.NoArgs),
		newTunnelCommand(cfg, "inspect", "inspect a managed tunnel", cobra.ExactArgs(1)),
		newTunnelCommand(cfg, "start", "start a managed tunnel", cobra.ExactArgs(1)),
		newTunnelCommand(cfg, "stop", "stop a managed tunnel", cobra.ExactArgs(1)),
		newTunnelCommand(cfg, "revoke", "revoke a managed tunnel", cobra.ExactArgs(1)),
		newSecretsCommand(cfg),
		newMachineTunnelCommand(cfg),
	)

	root.SetVersionTemplate("outpipe {{.Version}}\n")
	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print the outpipe version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(version)
			return nil
		},
	}
}
