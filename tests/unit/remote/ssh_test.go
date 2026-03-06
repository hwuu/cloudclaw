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

// TestWaitForSSH_ExponentialBackoff 测试指数退避逻辑
func TestWaitForSSH_ExponentialBackoff(t *testing.T) {
	var timestamps []time.Time
	mockDial := func() (remote.SSHClient, error) {
		timestamps = append(timestamps, time.Now())
		return nil, context.DeadlineExceeded
	}

	ctx := context.Background()
	_, err := remote.WaitForSSH(ctx, mockDial, remote.WaitSSHOptions{
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     200 * time.Millisecond,
		Timeout:         500 * time.Millisecond,
	})

	if err == nil {
		t.Fatal("WaitForSSH() error = nil, want timeout error")
	}

	// 验证至少重试了 3 次
	if len(timestamps) < 3 {
		t.Errorf("retry count = %d, want >= 3", len(timestamps))
	}

	// 验证指数退避：第二次间隔应大于第一次
	if len(timestamps) >= 3 {
		interval1 := timestamps[1].Sub(timestamps[0])
		interval2 := timestamps[2].Sub(timestamps[1])
		if interval2 < interval1 {
			t.Errorf("exponential backoff not working: interval1=%v, interval2=%v", interval1, interval2)
		}
	}
}

// TestWaitForSSH_ContextCancel 测试 context 取消场景
func TestWaitForSSH_ContextCancel(t *testing.T) {
	mockDial := func() (remote.SSHClient, error) {
		return nil, context.DeadlineExceeded
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	client, err := remote.WaitForSSH(ctx, mockDial, remote.WaitSSHOptions{
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Timeout:         1 * time.Second,
	})

	if err == nil {
		t.Fatal("WaitForSSH() error = nil, want context canceled error")
	}
	if client != nil {
		t.Errorf("WaitForSSH() client = %v, want nil", client)
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
