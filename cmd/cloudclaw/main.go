// Package main is the entry point for CloudClaw CLI.
// Provides subcommands: version (display version information).
// Version information is injected at build time via ldflags.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// Build-time injected via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cloudclaw",
		Short: "One-click deploy OpenClaw to Alibaba Cloud ECS",
		Long:  "CloudClaw - One-click deploy OpenClaw to Alibaba Cloud ECS with HTTPS and Gateway Token authentication.",
	}

	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cloudclaw %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
			fmt.Printf("  go:     %s\n", runtime.Version())
		},
	}
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
