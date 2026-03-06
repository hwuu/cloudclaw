// Package deploy 提供插件管理功能。
// 插件是 OpenClaw 的扩展模块，通过 SSH 远程安装/卸载。
package deploy

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hwuu/cloudclaw/internal/config"
	"github.com/hwuu/cloudclaw/internal/remote"
)

// PluginInfo 插件信息
type PluginInfo struct {
	Name        string // 插件名称
	Description string // 插件描述
	Version     string // 插件版本
	Enabled     bool   // 是否启用
	Installed   bool   // 是否已安装
}

// PluginManager 插件管理器
type PluginManager struct {
	Output      io.Writer
	SSHDialFunc SSHDialFactory
	StateDir    string // 覆盖默认 state 目录（测试用）
}

// 预定义插件列表
var knownPlugins = map[string]PluginInfo{
	"feishu": {
		Name:        "feishu",
		Description: "飞书消息通知插件",
		Version:     "latest",
	},
	"telegram": {
		Name:        "telegram",
		Description: "Telegram 消息通知插件",
		Version:     "latest",
	},
	"discord": {
		Name:        "discord",
		Description: "Discord 消息通知插件",
		Version:     "latest",
	},
	"wechat": {
		Name:        "wechat",
		Description: "企业微信消息通知插件",
		Version:     "latest",
	},
}

// ListPlugins 列出已安装插件
func (m *PluginManager) ListPlugins(ctx context.Context) ([]PluginInfo, error) {
	state, err := m.loadState()
	if err != nil {
		return nil, fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
	}

	privateKey, err := readSSHKeyFrom(m.getStateDir(), state)
	if err != nil {
		return nil, err
	}

	dialFunc := m.SSHDialFunc(state.Resources.EIP.IP, 22, "root", privateKey)
	sshClient, err := remote.WaitForSSH(ctx, dialFunc, remote.WaitSSHOptions{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败：%w", err)
	}
	defer sshClient.Close()

	// 获取已安装插件列表
	output, err := sshClient.RunCommand(ctx, `
		if [ -d /root/cloudclaw/plugins ]; then
			ls /root/cloudclaw/plugins
		fi
	`)
	if err != nil {
		return nil, fmt.Errorf("获取插件列表失败：%w", err)
	}

	var plugins []PluginInfo
	installedMap := make(map[string]bool)

	// 解析已安装插件
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			installedMap[line] = true
			if info, ok := knownPlugins[line]; ok {
				info.Installed = true
				info.Enabled = true // 默认已安装的插件都是启用的
				plugins = append(plugins, info)
			} else {
				plugins = append(plugins, PluginInfo{
					Name:      line,
					Installed: true,
					Enabled:   true,
				})
			}
		}
	}

	// 添加未安装但已知的插件
	for name, info := range knownPlugins {
		if !installedMap[name] {
			plugins = append(plugins, info)
		}
	}

	return plugins, nil
}

// InstallPlugin 安装插件
func (m *PluginManager) InstallPlugin(ctx context.Context, pluginName string) error {
	state, err := m.loadState()
	if err != nil {
		return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
	}

	// 检查是否是已知插件
	if _, ok := knownPlugins[pluginName]; !ok {
		return fmt.Errorf("未知插件：%s，支持的插件：%v", pluginName, m.getKnownPluginNames())
	}

	privateKey, err := readSSHKeyFrom(m.getStateDir(), state)
	if err != nil {
		return err
	}

	dialFunc := m.SSHDialFunc(state.Resources.EIP.IP, 22, "root", privateKey)
	sshClient, err := remote.WaitForSSH(ctx, dialFunc, remote.WaitSSHOptions{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("SSH 连接失败：%w", err)
	}
	defer sshClient.Close()

	m.printf("正在安装插件 %s...\n", pluginName)

	// 创建插件目录
	mkdirCmd := fmt.Sprintf("mkdir -p /root/cloudclaw/plugins/%s", pluginName)
	if _, err := sshClient.RunCommand(ctx, mkdirCmd); err != nil {
		return fmt.Errorf("创建插件目录失败：%w", err)
	}

	// 下载插件配置（从 GitHub 或预设配置）
	// 这里使用一个简单的占位配置，实际可以从远程下载
	configContent := m.getPluginConfig(pluginName)
	if configContent != "" {
		// 使用 heredoc 上传配置文件
		uploadCmd := fmt.Sprintf(`cat > /root/cloudclaw/plugins/%s/config.yml << 'EOF'
%s
EOF`, pluginName, configContent)
		if _, err := sshClient.RunCommand(ctx, uploadCmd); err != nil {
			return fmt.Errorf("上传插件配置失败：%w", err)
		}
	}

	// 标记插件为已启用
	touchCmd := fmt.Sprintf("touch /root/cloudclaw/plugins/%s/.enabled", pluginName)
	if _, err := sshClient.RunCommand(ctx, touchCmd); err != nil {
		return fmt.Errorf("启用插件失败：%w", err)
	}

	// 重启 Gateway 以加载插件
	m.printf("正在重启 Gateway 以加载插件...\n")
	if err := m.restartGateway(ctx, sshClient); err != nil {
		return fmt.Errorf("重启 Gateway 失败：%w", err)
	}

	m.printf("插件 %s 安装成功！\n", pluginName)
	return nil
}

// UninstallPlugin 卸载插件
func (m *PluginManager) UninstallPlugin(ctx context.Context, pluginName string) error {
	state, err := m.loadState()
	if err != nil {
		return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
	}

	// 检查插件是否已安装
	installed, err := m.isPluginInstalled(ctx, state, pluginName)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("插件 %s 未安装", pluginName)
	}

	privateKey, err := readSSHKeyFrom(m.getStateDir(), state)
	if err != nil {
		return err
	}

	dialFunc := m.SSHDialFunc(state.Resources.EIP.IP, 22, "root", privateKey)
	sshClient, err := remote.WaitForSSH(ctx, dialFunc, remote.WaitSSHOptions{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("SSH 连接失败：%w", err)
	}
	defer sshClient.Close()

	m.printf("正在卸载插件 %s...\n", pluginName)

	// 删除插件目录
	rmCmd := fmt.Sprintf("rm -rf /root/cloudclaw/plugins/%s", pluginName)
	if _, err := sshClient.RunCommand(ctx, rmCmd); err != nil {
		return fmt.Errorf("删除插件失败：%w", err)
	}

	// 重启 Gateway
	m.printf("正在重启 Gateway...\n")
	if err := m.restartGateway(ctx, sshClient); err != nil {
		return fmt.Errorf("重启 Gateway 失败：%w", err)
	}

	m.printf("插件 %s 卸载成功！\n", pluginName)
	return nil
}

// EnablePlugin 启用/禁用插件
func (m *PluginManager) EnablePlugin(ctx context.Context, pluginName string, enable bool) error {
	state, err := m.loadState()
	if err != nil {
		return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
	}

	// 检查插件是否已安装
	installed, err := m.isPluginInstalled(ctx, state, pluginName)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("插件 %s 未安装，请先安装", pluginName)
	}

	privateKey, err := readSSHKeyFrom(m.getStateDir(), state)
	if err != nil {
		return err
	}

	dialFunc := m.SSHDialFunc(state.Resources.EIP.IP, 22, "root", privateKey)
	sshClient, err := remote.WaitForSSH(ctx, dialFunc, remote.WaitSSHOptions{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("SSH 连接失败：%w", err)
	}
	defer sshClient.Close()

	enabledFile := fmt.Sprintf("/root/cloudclaw/plugins/%s/.enabled", pluginName)

	if enable {
		// 启用插件
		touchCmd := fmt.Sprintf("touch %s", enabledFile)
		if _, err := sshClient.RunCommand(ctx, touchCmd); err != nil {
			return fmt.Errorf("启用插件失败：%w", err)
		}
		m.printf("插件 %s 已启用\n", pluginName)
	} else {
		// 禁用插件
		rmCmd := fmt.Sprintf("rm -f %s", enabledFile)
		if _, err := sshClient.RunCommand(ctx, rmCmd); err != nil {
			return fmt.Errorf("禁用插件失败：%w", err)
		}
		m.printf("插件 %s 已禁用\n", pluginName)
	}

	// 重启 Gateway
	m.printf("正在重启 Gateway 以应用更改...\n")
	if err := m.restartGateway(ctx, sshClient); err != nil {
		return fmt.Errorf("重启 Gateway 失败：%w", err)
	}

	return nil
}

// --- 内部辅助方法 ---

func (m *PluginManager) printf(format string, args ...interface{}) {
	fmt.Fprintf(m.Output, format, args...)
}

func (m *PluginManager) getStateDir() string {
	if m.StateDir != "" {
		return m.StateDir
	}
	dir, _ := config.GetStateDir()
	return dir
}

func (m *PluginManager) loadState() (*config.State, error) {
	if m.StateDir != "" {
		return loadStateFrom(m.StateDir)
	}
	return config.LoadState()
}

func (m *PluginManager) getKnownPluginNames() []string {
	names := make([]string, 0, len(knownPlugins))
	for name := range knownPlugins {
		names = append(names, name)
	}
	return names
}

// getPluginConfig 获取插件配置模板
func (m *PluginManager) getPluginConfig(pluginName string) string {
	configs := map[string]string{
		"feishu": `# 飞书插件配置
feishu:
  webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
  secret: ""  # 可选，加签密钥`,
		"telegram": `# Telegram 插件配置
telegram:
  bot_token: ""  # Bot Token
  chat_id: ""    # 聊天 ID`,
		"discord": `# Discord 插件配置
discord:
  webhook_url: "https://discord.com/api/webhooks/xxx"`,
		"wechat": `# 企业微信插件配置
wechat:
  webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"`,
	}
	return configs[pluginName]
}

func (m *PluginManager) isPluginInstalled(ctx context.Context, state *config.State, pluginName string) (bool, error) {
	privateKey, err := readSSHKeyFrom(m.getStateDir(), state)
	if err != nil {
		return false, err
	}

	dialFunc := m.SSHDialFunc(state.Resources.EIP.IP, 22, "root", privateKey)
	sshClient, err := remote.WaitForSSH(ctx, dialFunc, remote.WaitSSHOptions{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return false, fmt.Errorf("SSH 连接失败：%w", err)
	}
	defer sshClient.Close()

	output, err := sshClient.RunCommand(ctx, fmt.Sprintf("ls /root/cloudclaw/plugins/%s 2>/dev/null && echo 'exists'", pluginName))
	if err != nil {
		return false, nil // 目录不存在，返回 false
	}
	return strings.Contains(output, "exists"), nil
}

func (m *PluginManager) restartGateway(ctx context.Context, sshClient remote.SSHClient) error {
	_, err := sshClient.RunCommand(ctx, "cd ~/cloudclaw && docker compose restart devbox")
	return err
}
