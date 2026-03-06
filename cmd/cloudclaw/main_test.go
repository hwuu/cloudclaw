package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

// TestNewRootCmd 测试根命令创建
func TestNewRootCmd(t *testing.T) {
	cmd := newRootCmd()

	if cmd == nil {
		t.Fatal("newRootCmd() returned nil")
	}

	if cmd.Use != "cloudclaw" {
		t.Errorf("newRootCmd() Use = %s, want 'cloudclaw'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("newRootCmd() Short should not be empty")
	}
}

// TestVersionCmd 测试 version 命令
func TestVersionCmd(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"version"})

	// version 命令使用 fmt.Printf 直接输出，不通过 cobra 缓冲
	// 只测试命令能成功执行
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

// TestDeployCmd_Help 测试 deploy 命令帮助信息
func TestDeployCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"deploy", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deploy --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "部署") {
		t.Errorf("deploy help should contain '部署', got: %s", output)
	}
}

// TestDestroyCmd_Help 测试 destroy 命令帮助信息
func TestDestroyCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"destroy", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("destroy --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "销毁") {
		t.Errorf("destroy help should contain '销毁', got: %s", output)
	}
}

// TestStatusCmd_Help 测试 status 命令帮助信息
func TestStatusCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"status", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("status --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "状态") {
		t.Errorf("status help should contain '状态', got: %s", output)
	}
}

// TestSuspendCmd_Help 测试 suspend 命令帮助信息
func TestSuspendCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"suspend", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("suspend --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "停止") {
		t.Errorf("suspend help should contain '停止', got: %s", output)
	}
}

// TestResumeCmd_Help 测试 resume 命令帮助信息
func TestResumeCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"resume", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("resume --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "恢复") {
		t.Errorf("resume help should contain '恢复', got: %s", output)
	}
}

// TestSSHCmd_Help 测试 ssh 命令帮助信息
func TestSSHCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"ssh", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("ssh --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "SSH") {
		t.Errorf("ssh help should contain 'SSH', got: %s", output)
	}
}

// TestExecCmd_Help 测试 exec 命令帮助信息
func TestExecCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"exec", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("exec --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "容器") {
		t.Errorf("exec help should contain '容器', got: %s", output)
	}
}

// TestPluginsCmd_Help 测试 plugins 命令帮助信息
func TestPluginsCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"plugins", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("plugins --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "插件") {
		t.Errorf("plugins help should contain '插件', got: %s", output)
	}
}

// TestChannelsCmd_Help 测试 channels 命令帮助信息
func TestChannelsCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"channels", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("channels --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "渠道") {
		t.Errorf("channels help should contain '渠道', got: %s", output)
	}
}

// TestRootCmd_NoArgs 测试无参数时显示帮助
func TestRootCmd_NoArgs(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command with no args failed: %v", err)
	}
}

// TestRootCmd_UnknownCommand 测试未知命令
func TestRootCmd_UnknownCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"unknown"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// 未知命令应该返回错误
	err := cmd.Execute()
	if err == nil {
		t.Fatal("unknown command should return error")
	}
}

// TestRootCmd_PersistentPreRunE_Version 测试 version 命令不需要配置
func TestRootCmd_PersistentPreRunE_Version(t *testing.T) {
	// 确保没有配置环境变量
	os.Unsetenv("ALICLOUD_ACCESS_KEY_ID")
	os.Unsetenv("ALICLOUD_ACCESS_KEY_SECRET")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"version"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command should not require config: %v", err)
	}
}

// TestRootCmd_PersistentPreRunE_Help 测试 help 命令不需要配置
func TestRootCmd_PersistentPreRunE_Help(t *testing.T) {
	// 确保没有配置环境变量
	os.Unsetenv("ALICLOUD_ACCESS_KEY_ID")
	os.Unsetenv("ALICLOUD_ACCESS_KEY_SECRET")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help command should not require config: %v", err)
	}
}

// TestPluginsListCmd_Help 测试 plugins list 命令帮助信息
func TestPluginsListCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"plugins", "list", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("plugins list --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "列出") {
		t.Errorf("plugins list help should contain '列出', got: %s", output)
	}
}

// TestPluginsInstallCmd_Help 测试 plugins install 命令帮助信息
func TestPluginsInstallCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"plugins", "install", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("plugins install --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "安装") {
		t.Errorf("plugins install help should contain '安装', got: %s", output)
	}
}

// TestChannelsListCmd_Help 测试 channels list 命令帮助信息
func TestChannelsListCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"channels", "list", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("channels list --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "列出") {
		t.Errorf("channels list help should contain '列出', got: %s", output)
	}
}

// TestChannelsAddCmd_Help 测试 channels add 命令帮助信息
func TestChannelsAddCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"channels", "add", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("channels add --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "添加") {
		t.Errorf("channels add help should contain '添加', got: %s", output)
	}
}

// TestRootCmd_SilenceUsage 测试 SilenceUsage 设置
func TestRootCmd_SilenceUsage(t *testing.T) {
	cmd := newRootCmd()

	if !cmd.SilenceUsage {
		t.Error("root command SilenceUsage should be true")
	}
}

// TestRootCmd_SilenceErrors 测试 SilenceErrors 设置
func TestRootCmd_SilenceErrors(t *testing.T) {
	cmd := newRootCmd()

	if !cmd.SilenceErrors {
		t.Error("root command SilenceErrors should be true")
	}
}

// TestDeployCmd_AppFlag 测试 deploy 命令的 --app 标志
func TestDeployCmd_AppFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"deploy", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deploy --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--app") {
		t.Errorf("deploy help should contain --app flag, got: %s", output)
	}
}

// TestDestroyCmd_ForceFlag 测试 destroy 命令的 --force 标志
func TestDestroyCmd_ForceFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"destroy", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("destroy --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--force") {
		t.Errorf("destroy help should contain --force flag, got: %s", output)
	}
}

// TestDestroyCmd_DryRunFlag 测试 destroy 命令的 --dry-run 标志
func TestDestroyCmd_DryRunFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"destroy", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("destroy --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--dry-run") {
		t.Errorf("destroy help should contain --dry-run flag, got: %s", output)
	}
}

// TestRootCmd_RegionFlag 测试根命令的 --region 标志
func TestRootCmd_RegionFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--region") {
		t.Errorf("root help should contain --region flag, got: %s", output)
	}
}

// TestExecCmd_ContainerFlag 测试 exec 命令的 --container 标志
func TestExecCmd_ContainerFlag(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"exec", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("exec --help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--container") {
		t.Errorf("exec help should contain --container flag, got: %s", output)
	}
}

// TestChannelsAddCmd_Flags 测试 channels add 命令的标志
func TestChannelsAddCmd_Flags(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"channels", "add", "--help"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("channels add --help failed: %v", err)
	}

	output := buf.String()
	requiredFlags := []string{"--type", "--webhook-url", "--bot-token", "--chat-id"}
	for _, flag := range requiredFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("channels add help should contain %s flag, got: %s", flag, output)
		}
	}
}

// TestContextCancellation 测试上下文取消
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	// 验证上下文确实被取消了
	select {
	case <-ctx.Done():
		// 预期行为
	default:
		t.Error("context should be cancelled")
	}
}
