package remote

import (
	"errors"
	"strings"
	"testing"

	"github.com/hwuu/cloudclaw/internal/remote"
)

// TestUploadFiles_Success 测试批量上传文件成功场景
func TestUploadFiles_Success(t *testing.T) {
	mockClient := &mockSFTPClient{
		files: make(map[string][]byte),
	}

	files := map[string][]byte{
		"/root/cloudclaw/docker-compose.yml": []byte("version: '3'"),
		"/root/cloudclaw/.env":               []byte("TOKEN=test"),
	}

	err := remote.UploadFiles(mockClient, files)
	if err != nil {
		t.Fatalf("UploadFiles() error = %v", err)
	}

	if len(mockClient.files) != 2 {
		t.Errorf("uploaded %d files, want 2", len(mockClient.files))
	}
}

// TestUploadFiles_Error 测试批量上传文件失败场景（任一失败立即返回）
func TestUploadFiles_Error(t *testing.T) {
	mockClient := &mockSFTPClient{
		files:       make(map[string][]byte),
		failOnPath:  "/root/cloudclaw/.env",
		failOnError: errors.New("disk full"),
	}

	files := map[string][]byte{
		"/root/cloudclaw/docker-compose.yml": []byte("version: '3'"),
		"/root/cloudclaw/.env":               []byte("TOKEN=test"),
	}

	err := remote.UploadFiles(mockClient, files)
	if err == nil {
		t.Fatal("UploadFiles() error = nil, want error")
	}

	// 验证错误消息包含期望内容
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("UploadFiles() error = %v, want to contain 'disk full'", err)
	}
}

// mockSFTPClient 模拟 SFTP 客户端
type mockSFTPClient struct {
	files       map[string][]byte
	failOnPath  string
	failOnError error
}

func (m *mockSFTPClient) UploadFile(content []byte, remotePath string) error {
	if m.failOnPath != "" && remotePath == m.failOnPath {
		return m.failOnError
	}
	m.files[remotePath] = content
	return nil
}

func (m *mockSFTPClient) Close() error {
	return nil
}
