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

// mockDeploySSHClient 模拟 SSH 客户端
type mockDeploySSHClient struct {
	failOnCommand bool
	commandOutput string
	// 支持根据不同命令返回不同输出
	commandOutputs map[string]string
}

func (m *mockDeploySSHClient) RunCommand(ctx context.Context, cmd string) (string, error) {
	if m.failOnCommand {
		return "", context.DeadlineExceeded
	}
	// 如果指定了命令级别的输出，优先使用
	if m.commandOutputs != nil {
		for pattern, output := range m.commandOutputs {
			if strings.Contains(cmd, pattern) {
				return output, nil
			}
		}
	}
	return m.commandOutput, nil
}

func (m *mockDeploySSHClient) Close() error {
	return nil
}

// setupTestDir 创建测试目录和文件
func setupTestDir(t *testing.T) string {
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

// TestPluginManager_ListPlugins_Success 测试列出插件成功场景
func TestPluginManager_ListPlugins_Success(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutputs: map[string]string{
			// 列出插件目录
			"ls /root/cloudclaw/plugins": "feishu\ntelegram",
			// 检查 .enabled 文件存在（feishu 已启用）
			"test -f /root/cloudclaw/plugins/feishu/.enabled": "yes",
			// 检查 .enabled 文件存在（telegram 已启用）
			"test -f /root/cloudclaw/plugins/telegram/.enabled": "yes",
		},
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	plugins, err := pm.ListPlugins(ctx)
	if err != nil {
		t.Fatalf("ListPlugins() error = %v", err)
	}

	// 验证返回了已知插件 + 已安装插件
	if len(plugins) < 2 {
		t.Errorf("ListPlugins() returned %d plugins, want >= 2", len(plugins))
	}

	// 验证已安装插件标记
	foundFeishu := false
	for _, p := range plugins {
		if p.Name == "feishu" && p.Installed {
			foundFeishu = true
			break
		}
	}
	if !foundFeishu {
		t.Error("ListPlugins() should mark feishu as installed")
	}
}

// TestPluginManager_ListPlugins_NoDeployment 测试无部署记录场景
func TestPluginManager_ListPlugins_NoDeployment(t *testing.T) {
	pm := &deploy.PluginManager{
		StateDir: t.TempDir(),
	}

	ctx := context.Background()
	_, err := pm.ListPlugins(ctx)
	if err == nil {
		t.Fatal("ListPlugins() error = nil, want error for no deployment")
	}
	if !strings.Contains(err.Error(), "未找到部署记录") {
		t.Errorf("ListPlugins() error = %v, want '未找到部署记录'", err)
	}
}

// TestPluginManager_InstallPlugin_Success 测试安装插件成功场景
func TestPluginManager_InstallPlugin_Success(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutputs: map[string]string{
			// 检查安装状态时返回空（未安装）
			"ls /root/cloudclaw/plugins/feishu": "",
			// 其他命令返回成功
			"mkdir":   "",
			"cat >":   "",
			"touch":   "",
			"docker":  "",
		},
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.InstallPlugin(ctx, "feishu")
	if err != nil {
		t.Fatalf("InstallPlugin() error = %v", err)
	}
}

// TestPluginManager_InstallPlugin_UnknownPlugin 测试安装未知插件场景
func TestPluginManager_InstallPlugin_UnknownPlugin(t *testing.T) {
	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output:   io.Discard,
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.InstallPlugin(ctx, "nonexistent")
	if err == nil {
		t.Fatal("InstallPlugin() error = nil, want error for unknown plugin")
	}
	if !strings.Contains(err.Error(), "未知插件") {
		t.Errorf("InstallPlugin() error = %v, want '未知插件'", err)
	}
}

// TestPluginManager_UninstallPlugin_Success 测试卸载插件成功场景
func TestPluginManager_UninstallPlugin_Success(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutputs: map[string]string{
			// 检查安装状态时返回存在
			"ls /root/cloudclaw/plugins/feishu": "exists",
			// 其他命令返回成功
			"rm":      "",
			"docker":  "",
		},
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.UninstallPlugin(ctx, "feishu")
	if err != nil {
		t.Fatalf("UninstallPlugin() error = %v", err)
	}
}

// TestPluginManager_UninstallPlugin_NotInstalled 测试卸载未安装插件场景
func TestPluginManager_UninstallPlugin_NotInstalled(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutput: "", // 空输出表示目录不存在
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.UninstallPlugin(ctx, "feishu")
	if err == nil {
		t.Fatal("UninstallPlugin() error = nil, want error for not installed plugin")
	}
	if !strings.Contains(err.Error(), "未安装") {
		t.Errorf("UninstallPlugin() error = %v, want '未安装'", err)
	}
}

// TestPluginManager_EnablePlugin_Success 测试启用插件成功场景
func TestPluginManager_EnablePlugin_Success(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutputs: map[string]string{
			// 检查安装状态时返回存在
			"ls /root/cloudclaw/plugins/feishu": "exists",
			// 其他命令返回成功
			"touch":  "",
			"docker": "",
		},
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.EnablePlugin(ctx, "feishu", true)
	if err != nil {
		t.Fatalf("EnablePlugin() error = %v", err)
	}
}

// TestPluginManager_EnablePlugin_NotInstalled 测试启用未安装插件场景
func TestPluginManager_EnablePlugin_NotInstalled(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutput: "", // 空输出表示目录不存在
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.EnablePlugin(ctx, "feishu", true)
	if err == nil {
		t.Fatal("EnablePlugin() error = nil, want error for not installed plugin")
	}
	if !strings.Contains(err.Error(), "未安装") {
		t.Errorf("EnablePlugin() error = %v, want '未安装'", err)
	}
}

// TestPluginManager_DisablePlugin_Success 测试禁用插件成功场景
func TestPluginManager_DisablePlugin_Success(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutputs: map[string]string{
			// 检查安装状态时返回存在
			"ls /root/cloudclaw/plugins/feishu": "exists",
			// 其他命令返回成功
			"rm":     "",
			"docker": "",
		},
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.EnablePlugin(ctx, "feishu", false)
	if err != nil {
		t.Fatalf("EnablePlugin(disable) error = %v", err)
	}
}

// TestPluginManager_SupportedPlugins 测试支持的插件列表
func TestPluginManager_SupportedPlugins(t *testing.T) {
	pm := &deploy.PluginManager{
		StateDir: t.TempDir(),
	}

	ctx := context.Background()
	supportedPlugins := []string{"feishu", "telegram", "discord", "wechat"}
	for _, name := range supportedPlugins {
		err := pm.InstallPlugin(ctx, name)
		// 预期会失败，因为无部署记录
		if err == nil {
			t.Errorf("InstallPlugin(%s) should fail without deployment", name)
		}
		if !strings.Contains(err.Error(), "未找到部署记录") {
			t.Errorf("InstallPlugin(%s) error = %v, want '未找到部署记录'", name, err)
		}
	}
}

// TestPluginManager_ListPlugins_SSHFailure 测试 SSH 连接失败场景
func TestPluginManager_ListPlugins_SSHFailure(t *testing.T) {
	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return nil, context.DeadlineExceeded
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	_, err := pm.ListPlugins(ctx)
	if err == nil {
		t.Fatal("ListPlugins() error = nil, want SSH failure error")
	}
	if !strings.Contains(err.Error(), "SSH 连接失败") {
		t.Errorf("ListPlugins() error = %v, want 'SSH 连接失败'", err)
	}
}

// TestPluginManager_InstallPlugin_AlreadyInstalled 测试重复安装场景
func TestPluginManager_InstallPlugin_AlreadyInstalled(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutputs: map[string]string{
			// 检查安装状态时返回存在（已安装）
			"ls /root/cloudclaw/plugins/feishu": "exists",
		},
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	err := pm.InstallPlugin(ctx, "feishu")
	if err == nil {
		t.Fatal("InstallPlugin() error = nil, want error for already installed plugin")
	}
	if !strings.Contains(err.Error(), "已安装") {
		t.Errorf("InstallPlugin() error = %v, want '已安装'", err)
	}
}

// TestPluginManager_ListPlugins_WithDisabledPlugin 测试列出插件包含禁用插件场景
func TestPluginManager_ListPlugins_WithDisabledPlugin(t *testing.T) {
	mockClient := &mockDeploySSHClient{
		commandOutputs: map[string]string{
			// 列出插件目录
			"ls /root/cloudclaw/plugins": "feishu\ntelegram",
			// 检查 .enabled 文件存在（feishu 已启用）
			"test -f /root/cloudclaw/plugins/feishu/.enabled": "yes",
			// 检查 .enabled 文件不存在（telegram 已禁用）
			"test -f /root/cloudclaw/plugins/telegram/.enabled": "",
		},
	}

	tempDir := setupTestDir(t)
	pm := &deploy.PluginManager{
		Output: io.Discard,
		SSHDialFunc: func(host string, port int, user string, privateKey []byte) remote.DialFunc {
			return func() (remote.SSHClient, error) {
				return mockClient, nil
			}
		},
		StateDir: tempDir,
	}

	ctx := context.Background()
	plugins, err := pm.ListPlugins(ctx)
	if err != nil {
		t.Fatalf("ListPlugins() error = %v", err)
	}

	// 验证 feishu 已启用
	foundEnabled := false
	for _, p := range plugins {
		if p.Name == "feishu" && p.Installed && p.Enabled {
			foundEnabled = true
			break
		}
	}
	if !foundEnabled {
		t.Error("ListPlugins() should mark feishu as enabled")
	}

	// 验证 telegram 已安装但禁用
	foundDisabled := false
	for _, p := range plugins {
		if p.Name == "telegram" && p.Installed && !p.Enabled {
			foundDisabled = true
			break
		}
	}
	if !foundDisabled {
		t.Error("ListPlugins() should mark telegram as disabled")
	}
}
