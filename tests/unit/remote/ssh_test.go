package remote

import (
	"context"
	"testing"
	"time"

	"github.com/hwuu/cloudclaw/internal/remote"
)

// TestWaitForSSH_Success 测试 SSH 连接成功场景
func TestWaitForSSH_Success(t *testing.T) {
	callCount := 0
	mockDial := func() (remote.SSHClient, error) {
		callCount++
		if callCount == 1 {
			return nil, context.DeadlineExceeded
		}
		return &mockSSHClient{}, nil
	}

	ctx := context.Background()
	client, err := remote.WaitForSSH(ctx, mockDial, remote.WaitSSHOptions{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Timeout:         1 * time.Second,
	})

	if err != nil {
		t.Fatalf("WaitForSSH() error = %v", err)
	}
	if client == nil {
		t.Fatal("WaitForSSH() client = nil, want non-nil")
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

// TestWaitForSSH_Timeout 测试 SSH 连接超时场景
func TestWaitForSSH_Timeout(t *testing.T) {
	mockDial := func() (remote.SSHClient, error) {
		return nil, context.DeadlineExceeded
	}

	ctx := context.Background()
	client, err := remote.WaitForSSH(ctx, mockDial, remote.WaitSSHOptions{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Timeout:         50 * time.Millisecond,
	})

	if err == nil {
		t.Fatal("WaitForSSH() error = nil, want timeout error")
	}
	if client != nil {
		t.Errorf("WaitForSSH() client = %v, want nil", client)
	}
}

// TestWaitForSSH_OptionsDefaults 测试选项默认值
func TestWaitForSSH_OptionsDefaults(t *testing.T) {
	opts := remote.WaitSSHOptions{}
	// 通过传递零值来测试默认值行为
	if opts.InitialInterval != 0 {
		t.Errorf("InitialInterval = %v, want 0 (zero value)", opts.InitialInterval)
	}
	if opts.MaxInterval != 0 {
		t.Errorf("MaxInterval = %v, want 0 (zero value)", opts.MaxInterval)
	}
	if opts.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0 (zero value)", opts.Timeout)
	}
}

// TestWaitForSSH_WithOptions 测试带选项的 WaitForSSH
func TestWaitForSSH_WithOptions(t *testing.T) {
	mockDial := func() (remote.SSHClient, error) {
		return &mockSSHClient{}, nil
	}

	ctx := context.Background()
	client, err := remote.WaitForSSH(ctx, mockDial, remote.WaitSSHOptions{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		Timeout:         1 * time.Second,
	})

	if err != nil {
		t.Fatalf("WaitForSSH() error = %v", err)
	}
	if client == nil {
		t.Fatal("WaitForSSH() client = nil, want non-nil")
	}
}

// mockSSHClient 模拟 SSH 客户端
type mockSSHClient struct{}

func (m *mockSSHClient) RunCommand(ctx context.Context, cmd string) (string, error) {
	return "ok", nil
}

func (m *mockSSHClient) Close() error {
	return nil
}
