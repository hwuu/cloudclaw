// Package main 是 CloudClaw CLI 的入口。
// 提供子命令：version（版本）。
// 版本信息通过 ldflags 在构建时注入。
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// 构建时通过 ldflags 注入
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cloudclaw",
		Short: "一键部署 OpenClaw 到阿里云 ECS",
		Long:  "CloudClaw — 一键部署 OpenClaw 到阿里云 ECS，带 HTTPS + Gateway Token 认证。",
	}

	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
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
