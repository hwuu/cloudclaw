// Package main 是 CloudClaw CLI 的入口。
// 提供子命令：deploy（部署）、destroy（销毁）、status（状态）、suspend（停机）、resume（恢复）、ssh（登录）、exec（容器命令）、plugins（插件管理）、version（版本）。
// 版本信息通过 ldflags 在构建时注入。
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hwuu/cloudclaw/internal/alicloud"
	"github.com/hwuu/cloudclaw/internal/config"
	"github.com/hwuu/cloudclaw/internal/deploy"
	"github.com/hwuu/cloudclaw/internal/remote"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// 构建时通过 ldflags 注入
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// 超时常量
const (
	deployTimeout  = 30 * time.Minute
	destroyTimeout = 15 * time.Minute
	defaultTimeout = 10 * time.Minute
	sshTimeout     = 10 * time.Second
	statusTimeout  = 5 * time.Minute
	execTimeout    = 10 * time.Minute
)

// 配置常量
const (
	defaultSSHUser = "root"
	defaultSSHPort = 22
)

// 全局标志
var (
	region  string
	force   bool
	dryRun  bool
	appOnly bool
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cloudclaw",
		Short: "一键部署 OpenClaw 到阿里云 ECS",
		Long:  "CloudClaw - 一键部署 OpenClaw 到阿里云 ECS，带 HTTPS + Gateway Token 认证。",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 需要配置的命令才检查配置
			if cmd.Name() == "version" || cmd.Name() == "help" || cmd.Name() == "__complete" {
				return nil
			}
			// 检查阿里云配置
			_, err := alicloud.LoadConfig()
			if err != nil {
				return fmt.Errorf("配置加载失败：%w\n请设置环境变量 ALICLOUD_ACCESS_KEY_ID 和 ALICLOUD_ACCESS_KEY_SECRET", err)
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&region, "region", "", "阿里云区域 (默认：ap-southeast-1)")
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newDeployCmd())
	rootCmd.AddCommand(newDestroyCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newSuspendCmd())
	rootCmd.AddCommand(newResumeCmd())
	rootCmd.AddCommand(newSSHCmd())
	rootCmd.AddCommand(newExecCmd())
	rootCmd.AddCommand(newPluginsCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("cloudclaw %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
			fmt.Printf("  go:     %s\n", runtime.Version())
		},
	}
}

// newDeployCmd 部署命令
func newDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "部署应用",
		Long:  "部署 CloudClaw 到阿里云 ECS，包括云资源创建和应用部署。",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), deployTimeout)
			defer cancel()

			clients, cfg, err := initClients()
			if err != nil {
				return err
			}

			prompter := config.NewDefaultPrompter()
			dialFunc := createSSHDialFunc()
			sftpFactory := createSFTPFactory()

			d := &deploy.Deployer{
				ECS:            clients.ECS,
				VPC:            clients.VPC,
				STS:            clients.STS,
				DNS:            clients.DNS,
				Prompter:       prompter,
				Output:         os.Stdout,
				Region:         cfg.RegionID,
				SSHDialFunc:    dialFunc,
				SFTPFactory:    sftpFactory,
				Version:        version,
				WaitInterval:   5 * time.Second,
				WaitTimeout:    5 * time.Minute,
				DNSWaitTimeout: 5 * time.Minute,
			}

			return d.Run(ctx, appOnly)
		},
	}
	cmd.Flags().BoolVar(&appOnly, "app", false, "仅重部署应用层（不创建云资源）")
	return cmd
}

// newDestroyCmd 销毁命令
func newDestroyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "销毁资源",
		Long:  "销毁 CloudClaw 创建的所有云资源，支持 --force（跳过确认）和 --dry-run（仅预览）。",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), destroyTimeout)
			defer cancel()

			clients, cfg, err := initClients()
			if err != nil {
				return err
			}

			prompter := config.NewDefaultPrompter()

			d := &deploy.Destroyer{
				ECS:          clients.ECS,
				VPC:          clients.VPC,
				Prompter:     prompter,
				Output:       os.Stdout,
				Region:       cfg.RegionID,
				Version:      version,
				WaitInterval: 5 * time.Second,
				WaitTimeout:  5 * time.Minute,
			}

			return d.Run(ctx, force, dryRun)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "跳过确认")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "仅预览，不实际删除")
	return cmd
}

// newStatusCmd 状态查询命令
func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "查询部署状态",
		Long:  "显示当前部署状态，包括云资源信息和容器运行状态。",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), statusTimeout)
			defer cancel()

			s := &deploy.StatusRunner{
				Output:      os.Stdout,
				SSHDialFunc: createSSHDialFunc(),
			}

			return s.Run(ctx)
		},
	}
}

// newSuspendCmd 停机命令
func newSuspendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend",
		Short: "停机",
		Long:  "停止 ECS 实例（StopCharging 模式），仅收取磁盘费用。",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), defaultTimeout)
			defer cancel()

			clients, cfg, err := initClients()
			if err != nil {
				return err
			}

			prompter := config.NewDefaultPrompter()

			s := &deploy.Suspender{
				ECS:          clients.ECS,
				Prompter:     prompter,
				Output:       os.Stdout,
				Region:       cfg.RegionID,
				WaitInterval: 5 * time.Second,
				WaitTimeout:  5 * time.Minute,
			}

			return s.Run(ctx)
		},
	}
	return cmd
}

// newResumeCmd 恢复命令
func newResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "恢复运行",
		Long:  "启动已停止的 ECS 实例，恢复服务运行。",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), defaultTimeout)
			defer cancel()

			clients, cfg, err := initClients()
			if err != nil {
				return err
			}

			prompter := config.NewDefaultPrompter()

			r := &deploy.Resumer{
				ECS:          clients.ECS,
				Prompter:     prompter,
				Output:       os.Stdout,
				Region:       cfg.RegionID,
				SSHDialFunc:  createSSHDialFunc(),
				WaitInterval: 5 * time.Second,
				WaitTimeout:  5 * time.Minute,
			}

			return r.Run(ctx)
		},
	}
	return cmd
}

// connectSSH 连接到 ECS 实例（公共函数，避免代码重复）
func connectSSH(state *config.State) (*ssh.Client, error) {
	// 检查 EIP 是否存在
	if state.Resources.EIP.ID == "" || state.Resources.EIP.IP == "" {
		return nil, fmt.Errorf("ECS 实例未创建或无公网 IP")
	}

	stateDir, err := config.GetStateDir()
	if err != nil {
		return nil, err
	}

	privateKey, err := os.ReadFile(filepath.Join(stateDir, "ssh_key"))
	if err != nil {
		return nil, fmt.Errorf("读取 SSH 私钥失败：%w", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("解析 SSH 私钥失败：%w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: defaultSSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         sshTimeout,
	}

	addr := fmt.Sprintf("%s:%d", state.Resources.EIP.IP, defaultSSHPort)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败：%w", err)
	}

	return client, nil
}

// newSSHCmd SSH 登录命令
func newSSHCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ssh",
		Short: "SSH 登录 ECS 实例",
		Long:  "通过 SSH 连接到 CloudClaw 创建的 ECS 实例。",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := config.LoadState()
			if err != nil {
				return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
			}

			client, err := connectSSH(state)
			if err != nil {
				return err
			}
			defer client.Close()

			session, err := client.NewSession()
			if err != nil {
				return fmt.Errorf("创建 SSH 会话失败：%w", err)
			}
			defer session.Close()

			// 设置终端模式
			session.Stdin = os.Stdin
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr

			// 请求伪终端
			modes := ssh.TerminalModes{
				ssh.ECHO:          1,
				ssh.TTY_OP_ISPEED: 14400,
				ssh.TTY_OP_OSPEED: 14400,
			}

			if err := session.RequestPty("xterm", 80, 24, modes); err != nil {
				return fmt.Errorf("请求伪终端失败：%w", err)
			}

			if err := session.Shell(); err != nil {
				return fmt.Errorf("启动 SSH Shell 失败：%w", err)
			}

			// 等待会话结束
			if err := session.Wait(); err != nil {
				if exitErr, ok := err.(*ssh.ExitError); ok {
					return fmt.Errorf("SSH 会话退出码：%d", exitErr.ExitStatus())
				}
				return err
			}

			return nil
		},
	}
}

// newExecCmd 容器命令执行命令
func newExecCmd() *cobra.Command {
	var container string
	cmd := &cobra.Command{
		Use:   "exec <command>",
		Short: "在容器中执行命令",
		Long:  "在指定的 Docker 容器中执行命令，默认容器为 devbox。",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := config.LoadState()
			if err != nil {
				return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
			}

			client, err := connectSSH(state)
			if err != nil {
				return err
			}
			defer client.Close()

			session, err := client.NewSession()
			if err != nil {
				return fmt.Errorf("创建 SSH 会话失败：%w", err)
			}
			defer session.Close()

			// 设置超时
			ctx, cancel := context.WithTimeout(cmd.Context(), execTimeout)
			defer cancel()

			// 构建 docker exec 命令
			execCmd := fmt.Sprintf("docker exec -it %s %s", container, strings.Join(args, " "))
			session.Stdin = os.Stdin
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr

			done := make(chan error, 1)
			go func() {
				done <- session.Run(execCmd)
			}()

			select {
			case <-ctx.Done():
				_ = session.Signal(ssh.SIGTERM)
				<-done
				return fmt.Errorf("命令执行超时：%w", ctx.Err())
			case err := <-done:
				if err != nil {
					if exitErr, ok := err.(*ssh.ExitError); ok {
						return fmt.Errorf("命令退出码：%d", exitErr.ExitStatus())
					}
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&container, "container", "devbox", "容器名称")
	return cmd
}

// initClients 初始化阿里云客户端
func initClients() (*alicloud.Clients, *alicloud.Config, error) {
	cfg, err := alicloud.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	if region != "" {
		cfg.RegionID = region
	}

	clients, err := alicloud.NewClients(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("初始化阿里云客户端失败：%w", err)
	}

	return clients, cfg, nil
}

// createSSHDialFunc 创建 SSH DialFunc
func createSSHDialFunc() deploy.SSHDialFactory {
	return func(host string, port int, user string, privateKey []byte) remote.DialFunc {
		return func() (remote.SSHClient, error) {
			signer, err := ssh.ParsePrivateKey(privateKey)
			if err != nil {
				return nil, err
			}

			sshConfig := &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Timeout:         sshTimeout,
			}

			addr := fmt.Sprintf("%s:%d", host, port)
			client, err := ssh.Dial("tcp", addr, sshConfig)
			if err != nil {
				return nil, err
			}

			return &sshClientWrapper{client: client}, nil
		}
	}
}

// createSFTPFactory 创建 SFTP 工厂函数
func createSFTPFactory() deploy.SFTPClientFactory {
	return func(host string, port int, user string, privateKey []byte) (remote.SFTPClient, error) {
		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, err
		}

		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         sshTimeout,
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		client, err := ssh.Dial("tcp", addr, sshConfig)
		if err != nil {
			return nil, err
		}

		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			client.Close()
			return nil, err
		}

		return &sftpClientWrapper{
			sshClient:  client,
			sftpClient: sftpClient,
		}, nil
	}
}

// sshClientWrapper SSH 客户端包装器
type sshClientWrapper struct {
	client *ssh.Client
}

func (w *sshClientWrapper) RunCommand(ctx context.Context, cmd string) (string, error) {
	session, err := w.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var output string
	var runErr error

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf, err := session.CombinedOutput(cmd)
		output = string(buf)
		runErr = err
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		<-done
		return "", ctx.Err()
	case <-done:
		return output, runErr
	}
}

func (w *sshClientWrapper) Close() error {
	return w.client.Close()
}

// sftpClientWrapper SFTP 客户端包装器
type sftpClientWrapper struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func (w *sftpClientWrapper) UploadFile(content []byte, remotePath string) error {
	dir := filepath.Dir(remotePath)
	if err := w.sftpClient.MkdirAll(dir); err != nil {
		return fmt.Errorf("创建远程目录 %s 失败：%w", dir, err)
	}

	f, err := w.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件 %s 失败：%w", remotePath, err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("写入远程文件 %s 失败：%w", remotePath, err)
	}

	return nil
}

func (w *sftpClientWrapper) Close() error {
	sftpErr := w.sftpClient.Close()
	sshErr := w.sshClient.Close()
	if sftpErr != nil {
		return sftpErr
	}
	return sshErr
}

// newPluginsCmd 插件管理命令
func newPluginsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "插件管理",
		Long:  "管理 OpenClaw 插件，支持安装、卸载、启用/禁用插件。",
	}

	cmd.AddCommand(newPluginsListCmd())
	cmd.AddCommand(newPluginsInstallCmd())
	cmd.AddCommand(newPluginsUninstallCmd())
	cmd.AddCommand(newPluginsEnableCmd())
	cmd.AddCommand(newPluginsDisableCmd())

	return cmd
}

// newPluginsListCmd 列出插件命令
func newPluginsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出插件",
		Long:  "列出所有可用插件及安装状态。",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			pm := &deploy.PluginManager{
				Output:      os.Stdout,
				SSHDialFunc: createSSHDialFunc(),
			}

			plugins, err := pm.ListPlugins(ctx)
			if err != nil {
				return err
			}

			if len(plugins) == 0 {
				fmt.Println("暂无插件")
				return nil
			}

			fmt.Printf("%-12s %-30s %-10s %-8s\n", "名称", "描述", "版本", "状态")
			fmt.Println(strings.Repeat("-", 70))
			for _, p := range plugins {
				status := "未安装"
				if p.Installed {
					if p.Enabled {
						status = "已启用"
					} else {
						status = "已禁用"
					}
				}
				fmt.Printf("%-12s %-30s %-10s %-8s\n", p.Name, p.Description, p.Version, status)
			}
			return nil
		},
	}
}

// newPluginsInstallCmd 安装插件命令
func newPluginsInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <plugin>",
		Short: "安装插件",
		Long:  "安装指定的插件，支持 feishu、telegram、discord、wechat 等。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			pluginName := args[0]

			pm := &deploy.PluginManager{
				Output:      os.Stdout,
				SSHDialFunc: createSSHDialFunc(),
			}

			return pm.InstallPlugin(ctx, pluginName)
		},
	}
}

// newPluginsUninstallCmd 卸载插件命令
func newPluginsUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <plugin>",
		Short: "卸载插件",
		Long:  "卸载指定的已安装插件。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			pluginName := args[0]

			pm := &deploy.PluginManager{
				Output:      os.Stdout,
				SSHDialFunc: createSSHDialFunc(),
			}

			return pm.UninstallPlugin(ctx, pluginName)
		},
	}
}

// newPluginsEnableCmd 启用插件命令
func newPluginsEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <plugin>",
		Short: "启用插件",
		Long:  "启用指定的已安装插件。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			pluginName := args[0]

			pm := &deploy.PluginManager{
				Output:      os.Stdout,
				SSHDialFunc: createSSHDialFunc(),
			}

			return pm.EnablePlugin(ctx, pluginName, true)
		},
	}
}

// newPluginsDisableCmd 禁用插件命令
func newPluginsDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <plugin>",
		Short: "禁用插件",
		Long:  "禁用指定的已安装插件。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			pluginName := args[0]

			pm := &deploy.PluginManager{
				Output:      os.Stdout,
				SSHDialFunc: createSSHDialFunc(),
			}

			return pm.EnablePlugin(ctx, pluginName, false)
		},
	}
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
