package deploy

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/hwuu/cloudclaw/internal/config"
	"github.com/hwuu/cloudclaw/internal/deploy"
	"github.com/hwuu/cloudclaw/internal/remote"
)

// mockChannelSSHClient 模拟渠道测试 SSH 客户端
type mockChannelSSHClient struct {
	commandOutput string
	failOnCommand bool
}

func (m *mockChannelSSHClient) RunCommand(ctx context.Context, cmd string) (string, error) {
	if m.failOnCommand {
		return "", context.DeadlineExceeded
	}
	return m.commandOutput, nil
}

func (m *mockChannelSSHClient) Close() error {
	return nil
}

// setupChannelTestDir 创建渠道测试目录和文件
func setupChannelTestDir(t *testing.T) string {
	tempDir := t.TempDir()
	state := &config.State{
		Version: config.StateFileVersion,
		Region:  "ap-southeast-1",
		Resources: config.Resources{
			ECS: config.ECSResource{ID: "i-test"},
			EIP: config.EIPResource{ID: "eip-test", IP: "1.2.3.4"},
		},
	}
	saveTestStateHelper(tempDir, state)
	saveTestSSHKeyHelper(tempDir)
	return tempDir
}

// TestChannelManager_ListChannels_Success 测试列出渠道成功场景
func TestChannelManager_ListChannels_Success(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: `[
  {
    "name": "test-feishu",
    "type": "feishu",
    "enabled": true,
    "config": "{\"webhook_url\":\"https://example.com\"}"
  }
]`,
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	channels, err := cm.ListChannels(ctx)
	if err != nil {
		t.Fatalf("ListChannels() error = %v", err)
	}

	if len(channels) != 1 {
		t.Errorf("ListChannels() returned %d channels, want 1", len(channels))
	}

	if channels[0].Name != "test-feishu" {
		t.Errorf("ListChannels()[0].Name = %s, want 'test-feishu'", channels[0].Name)
	}
}

// TestChannelManager_ListChannels_Empty 测试空渠道列表
func TestChannelManager_ListChannels_Empty(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: "",
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	channels, err := cm.ListChannels(ctx)
	if err != nil {
		t.Fatalf("ListChannels() error = %v", err)
	}

	if len(channels) != 0 {
		t.Errorf("ListChannels() returned %d channels, want 0", len(channels))
	}
}

// TestChannelManager_ListChannels_NoDeployment 测试无部署记录场景
func TestChannelManager_ListChannels_NoDeployment(t *testing.T) {
	cm := &deploy.ChannelManager{
		StateDir: t.TempDir(),
	}

	ctx := context.Background()
	_, err := cm.ListChannels(ctx)
	if err == nil {
		t.Fatal("ListChannels() error = nil, want error for no deployment")
	}
	if !strings.Contains(err.Error(), "未找到部署记录") {
		t.Errorf("ListChannels() error = %v, want '未找到部署记录'", err)
	}
}

// TestChannelManager_AddChannel_Success 测试添加渠道成功场景
func TestChannelManager_AddChannel_Success(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: "",
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	cfg := deploy.ChannelConfig{
		WebhookURL: "https://example.com/webhook",
	}

	err := cm.AddChannel(ctx, "test-feishu", "feishu", cfg)
	if err != nil {
		t.Fatalf("AddChannel() error = %v", err)
	}
}

// TestChannelManager_AddChannel_UnknownType 测试添加未知渠道类型
func TestChannelManager_AddChannel_UnknownType(t *testing.T) {
	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return &mockChannelSSHClient{}, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	cfg := deploy.ChannelConfig{}

	err := cm.AddChannel(ctx, "test", "unknown", cfg)
	if err == nil {
		t.Fatal("AddChannel() error = nil, want error for unknown type")
	}
	if !strings.Contains(err.Error(), "未知渠道类型") {
		t.Errorf("AddChannel() error = %v, want '未知渠道类型'", err)
	}
}

// TestChannelManager_RemoveChannel_Success 测试删除渠道成功场景
func TestChannelManager_RemoveChannel_Success(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: `[
  {
    "name": "test-feishu",
    "type": "feishu",
    "enabled": true,
    "config": "{\"webhook_url\":\"https://example.com\"}"
  }
]`,
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := cm.RemoveChannel(ctx, "test-feishu")
	if err != nil {
		t.Fatalf("RemoveChannel() error = %v", err)
	}
}

// TestChannelManager_RemoveChannel_NotFound 测试删除不存在的渠道
func TestChannelManager_RemoveChannel_NotFound(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: "[]",
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := cm.RemoveChannel(ctx, "nonexistent")
	if err == nil {
		t.Fatal("RemoveChannel() error = nil, want error for not found")
	}
	if !strings.Contains(err.Error(), "不存在") {
		t.Errorf("RemoveChannel() error = %v, want '不存在'", err)
	}
}

// TestChannelManager_EnableChannel_Success 测试启用渠道成功场景
func TestChannelManager_EnableChannel_Success(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: `[
  {
    "name": "test-feishu",
    "type": "feishu",
    "enabled": false,
    "config": "{\"webhook_url\":\"https://example.com\"}"
  }
]`,
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := cm.EnableChannel(ctx, "test-feishu", true)
	if err != nil {
		t.Fatalf("EnableChannel() error = %v", err)
	}
}

// TestChannelManager_EnableChannel_NotFound 测试启用不存在的渠道
func TestChannelManager_EnableChannel_NotFound(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: "[]",
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := cm.EnableChannel(ctx, "nonexistent", true)
	if err == nil {
		t.Fatal("EnableChannel() error = nil, want error for not found")
	}
	if !strings.Contains(err.Error(), "不存在") {
		t.Errorf("EnableChannel() error = %v, want '不存在'", err)
	}
}

// TestChannelManager_DisableChannel_Success 测试禁用渠道成功场景
func TestChannelManager_DisableChannel_Success(t *testing.T) {
	mockClient := &mockChannelSSHClient{
		commandOutput: `[
  {
    "name": "test-feishu",
    "type": "feishu",
    "enabled": true,
    "config": "{\"webhook_url\":\"https://example.com\"}"
  }
]`,
	}

	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := cm.EnableChannel(ctx, "test-feishu", false)
	if err != nil {
		t.Fatalf("EnableChannel(disable) error = %v", err)
	}
}

// TestChannelManager_KnownChannelTypes 测试已知渠道类型
func TestChannelManager_KnownChannelTypes(t *testing.T) {
	expectedTypes := []string{"feishu", "telegram", "discord", "wechat"}

	cm := &deploy.ChannelManager{
		StateDir: t.TempDir(),
	}

	ctx := context.Background()
	for _, channelType := range expectedTypes {
		err := cm.AddChannel(ctx, "test", channelType, deploy.ChannelConfig{})
		// 预期会失败，因为无部署记录
		if err == nil {
			t.Errorf("AddChannel(%s) should fail without deployment", channelType)
		}
		if !strings.Contains(err.Error(), "未找到部署记录") {
			t.Errorf("AddChannel(%s) error = %v, want '未找到部署记录'", channelType, err)
		}
	}
}

// TestChannelManager_ListChannels_SSHFailure 测试 SSH 连接失败场景
func TestChannelManager_ListChannels_SSHFailure(t *testing.T) {
	tempDir := setupChannelTestDir(t)
	cm := &deploy.ChannelManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return nil, context.DeadlineExceeded
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	_, err := cm.ListChannels(ctx)
	if err == nil {
		t.Fatal("ListChannels() error = nil, want SSH failure error")
	}
	if !strings.Contains(err.Error(), "SSH 连接失败") {
		t.Errorf("ListChannels() error = %v, want 'SSH 连接失败'", err)
	}
}
