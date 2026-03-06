// Package deploy 提供渠道管理功能。
// 渠道是 OpenClaw 的消息通知目标，通过 SSH 远程配置。
package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hwuu/cloudclaw/internal/config"
	"github.com/hwuu/cloudclaw/internal/remote"
)

// ChannelInfo 渠道信息
type ChannelInfo struct {
	Name     string // 渠道名称
	Type     string // 渠道类型 (feishu/telegram/discord/wechat)
	Enabled  bool   // 是否启用
	Config   string // 渠道配置 (JSON 格式)
}

// ChannelConfig 渠道配置结构
type ChannelConfig struct {
	WebhookURL string `json:"webhook_url,omitempty"`
	BotToken   string `json:"bot_token,omitempty"`
	ChatID     string `json:"chat_id,omitempty"`
	Secret     string `json:"secret,omitempty"`
}

// ChannelManager 渠道管理器
type ChannelManager struct {
	Output      io.Writer
	SSHDialFunc SSHDialFactory
	StateDir    string // 覆盖默认 state 目录（测试用）
}

// 预定义渠道类型
var knownChannelTypes = map[string]string{
	"feishu":   "飞书机器人",
	"telegram": "Telegram Bot",
	"discord":  "Discord Webhook",
	"wechat":   "企业微信机器人",
}

// ListChannels 列出已配置渠道
func (m *ChannelManager) ListChannels(ctx context.Context) ([]ChannelInfo, error) {
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

	// 读取渠道配置文件
	output, err := sshClient.RunCommand(ctx, "cat /root/cloudclaw/channels.json 2>/dev/null || echo ''")
	if err != nil {
		return nil, fmt.Errorf("读取渠道配置失败：%w", err)
	}

	if strings.TrimSpace(output) == "" {
		return []ChannelInfo{}, nil
	}

	// 解析渠道配置
	var channels []ChannelInfo
	if err := json.Unmarshal([]byte(output), &channels); err != nil {
		return nil, fmt.Errorf("解析渠道配置失败：%w", err)
	}

	return channels, nil
}

// AddChannel 添加渠道
func (m *ChannelManager) AddChannel(ctx context.Context, name, channelType string, cfg ChannelConfig) error {
	state, err := m.loadState()
	if err != nil {
		return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
	}

	// 检查渠道类型
	if _, ok := knownChannelTypes[channelType]; !ok {
		return fmt.Errorf("未知渠道类型：%s，支持的类型：%v", channelType, m.getKnownChannelTypes())
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

	m.printf("正在添加渠道 %s (%s)...\n", name, channelType)

	// 读取现有渠道配置
	output, err := sshClient.RunCommand(ctx, "cat /root/cloudclaw/channels.json 2>/dev/null || echo '[]'")
	if err != nil {
		return fmt.Errorf("读取渠道配置失败：%w", err)
	}

	var channels []ChannelInfo
	if strings.TrimSpace(output) != "" {
		if err := json.Unmarshal([]byte(output), &channels); err != nil {
			return fmt.Errorf("解析渠道配置失败：%w", err)
		}
	}

	// 检查是否已存在同名渠道
	for _, ch := range channels {
		if ch.Name == name {
			return fmt.Errorf("渠道 %s 已存在", name)
		}
	}

	// 序列化配置
	configJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败：%w", err)
	}

	// 添加新渠道
	channels = append(channels, ChannelInfo{
		Name:    name,
		Type:    channelType,
		Enabled: true,
		Config:  string(configJSON),
	})

	// 写回配置文件
	channelsJSON, err := json.MarshalIndent(channels, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化渠道列表失败：%w", err)
	}

	// 使用 heredoc 上传配置文件
	uploadCmd := fmt.Sprintf(`cat > /root/cloudclaw/channels.json << 'EOF'
%s
EOF`, string(channelsJSON))
	if _, err := sshClient.RunCommand(ctx, uploadCmd); err != nil {
		return fmt.Errorf("写入渠道配置失败：%w", err)
	}

	// 重启 Gateway
	m.printf("正在重启 Gateway 以加载配置...\n")
	if err := m.restartGateway(ctx, sshClient); err != nil {
		return fmt.Errorf("重启 Gateway 失败：%w", err)
	}

	m.printf("渠道 %s 添加成功！\n", name)
	return nil
}

// RemoveChannel 删除渠道
func (m *ChannelManager) RemoveChannel(ctx context.Context, name string) error {
	state, err := m.loadState()
	if err != nil {
		return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
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

	m.printf("正在删除渠道 %s...\n", name)

	// 读取现有渠道配置
	output, err := sshClient.RunCommand(ctx, "cat /root/cloudclaw/channels.json 2>/dev/null || echo '[]'")
	if err != nil {
		return fmt.Errorf("读取渠道配置失败：%w", err)
	}

	var channels []ChannelInfo
	if strings.TrimSpace(output) != "" {
		if err := json.Unmarshal([]byte(output), &channels); err != nil {
			return fmt.Errorf("解析渠道配置失败：%w", err)
		}
	}

	// 查找并删除渠道
	found := false
	newChannels := make([]ChannelInfo, 0, len(channels))
	for _, ch := range channels {
		if ch.Name == name {
			found = true
		} else {
			newChannels = append(newChannels, ch)
		}
	}

	if !found {
		return fmt.Errorf("渠道 %s 不存在", name)
	}

	// 写回配置文件
	channelsJSON, err := json.MarshalIndent(newChannels, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化渠道列表失败：%w", err)
	}

	// 使用 heredoc 上传配置文件
	uploadCmd := fmt.Sprintf(`cat > /root/cloudclaw/channels.json << 'EOF'
%s
EOF`, string(channelsJSON))
	if _, err := sshClient.RunCommand(ctx, uploadCmd); err != nil {
		return fmt.Errorf("写入渠道配置失败：%w", err)
	}

	// 重启 Gateway
	m.printf("正在重启 Gateway 以加载配置...\n")
	if err := m.restartGateway(ctx, sshClient); err != nil {
		return fmt.Errorf("重启 Gateway 失败：%w", err)
	}

	m.printf("渠道 %s 删除成功！\n", name)
	return nil
}

// EnableChannel 启用/禁用渠道
func (m *ChannelManager) EnableChannel(ctx context.Context, name string, enable bool) error {
	state, err := m.loadState()
	if err != nil {
		return fmt.Errorf("未找到部署记录，请先运行 cloudclaw deploy")
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

	// 读取现有渠道配置
	output, err := sshClient.RunCommand(ctx, "cat /root/cloudclaw/channels.json 2>/dev/null || echo '[]'")
	if err != nil {
		return fmt.Errorf("读取渠道配置失败：%w", err)
	}

	var channels []ChannelInfo
	if strings.TrimSpace(output) != "" {
		if err := json.Unmarshal([]byte(output), &channels); err != nil {
			return fmt.Errorf("解析渠道配置失败：%w", err)
		}
	}

	// 查找并更新渠道状态
	found := false
	for i, ch := range channels {
		if ch.Name == name {
			channels[i].Enabled = enable
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("渠道 %s 不存在", name)
	}

	// 写回配置文件
	channelsJSON, err := json.MarshalIndent(channels, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化渠道列表失败：%w", err)
	}

	// 使用 heredoc 上传配置文件
	uploadCmd := fmt.Sprintf(`cat > /root/cloudclaw/channels.json << 'EOF'
%s
EOF`, string(channelsJSON))
	if _, err := sshClient.RunCommand(ctx, uploadCmd); err != nil {
		return fmt.Errorf("写入渠道配置失败：%w", err)
	}

	// 重启 Gateway
	m.printf("正在重启 Gateway 以加载配置...\n")
	if err := m.restartGateway(ctx, sshClient); err != nil {
		return fmt.Errorf("重启 Gateway 失败：%w", err)
	}

	if enable {
		m.printf("渠道 %s 已启用\n", name)
	} else {
		m.printf("渠道 %s 已禁用\n", name)
	}
	return nil
}

// --- 内部辅助方法 ---

func (m *ChannelManager) printf(format string, args ...interface{}) {
	fmt.Fprintf(m.Output, format, args...)
}

func (m *ChannelManager) getStateDir() string {
	if m.StateDir != "" {
		return m.StateDir
	}
	dir, _ := config.GetStateDir()
	return dir
}

func (m *ChannelManager) loadState() (*config.State, error) {
	if m.StateDir != "" {
		return loadStateFrom(m.StateDir)
	}
	return config.LoadState()
}

func (m *ChannelManager) getKnownChannelTypes() []string {
	types := make([]string, 0, len(knownChannelTypes))
	for t := range knownChannelTypes {
		types = append(types, t)
	}
	return types
}

func (m *ChannelManager) restartGateway(ctx context.Context, sshClient remote.SSHClient) error {
	_, err := sshClient.RunCommand(ctx, "cd ~/cloudclaw && docker compose restart devbox")
	return err
}
