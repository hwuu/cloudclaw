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

// TestUploadFiles_EmptyMap 测试空文件 map 边界情况
func TestUploadFiles_EmptyMap(t *testing.T) {
	mockClient := &mockSFTPClient{
		files: make(map[string][]byte),
	}

	files := map[string][]byte{}

	err := remote.UploadFiles(mockClient, files)
	if err != nil {
		t.Fatalf("UploadFiles() error = %v, want nil", err)
	}

	if len(mockClient.files) != 0 {
		t.Errorf("uploaded %d files, want 0", len(mockClient.files))
	}
}

// TestUploadFiles_NilMap 测试 nil map 边界情况
func TestUploadFiles_NilMap(t *testing.T) {
	mockClient := &mockSFTPClient{
		files: make(map[string][]byte),
	}

	var files map[string][]byte // nil map

	err := remote.UploadFiles(mockClient, files)
	if err != nil {
		t.Fatalf("UploadFiles(nil) error = %v, want nil", err)
	}

	if len(mockClient.files) != 0 {
		t.Errorf("uploaded %d files, want 0", len(mockClient.files))
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

// TestUploadFiles_ErrorStopsUpload 测试错误发生后停止上传（无部分上传）
func TestUploadFiles_ErrorStopsUpload(t *testing.T) {
	mockClient := &mockSFTPClient{
		files:       make(map[string][]byte),
		failOnPath:  "/root/cloudclaw/.env",
		failOnError: errors.New("disk full"),
	}

	// 只有会失败的文件
	files := map[string][]byte{
		"/root/cloudclaw/.env": []byte("TOKEN=test"),
	}

	err := remote.UploadFiles(mockClient, files)
	if err == nil {
		t.Fatal("UploadFiles() error = nil, want error")
	}

	// 验证失败后没有文件被上传
	if len(mockClient.files) != 0 {
		t.Errorf("uploaded %d files after error, want 0", len(mockClient.files))
	}
}

// TestUploadFiles_MultipleFilesWithMultipleErrors 测试多文件场景下第一个失败即停止
func TestUploadFiles_MultipleFilesWithMultipleErrors(t *testing.T) {
	mockClient := &mockSFTPClient{
		files:       make(map[string][]byte),
		failOnPath:  "/root/cloudclaw/.env",
		failOnError: errors.New("disk full"),
	}

	// 多个文件，其中一个会失败
	files := map[string][]byte{
		"/root/cloudclaw/docker-compose.yml": []byte("version: '3'"),
		"/root/cloudclaw/.env":               []byte("TOKEN=test"),
		"/root/cloudclaw/Caddyfile":          []byte("localhost"),
	}

	err := remote.UploadFiles(mockClient, files)
	if err == nil {
		t.Fatal("UploadFiles() error = nil, want error")
	}

	// 验证错误消息
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("UploadFiles() error = %v, want to contain 'disk full'", err)
	}

	// 验证部分上传：在错误前的文件可能已上传（map 遍历顺序不确定）
	// 但至少验证失败文件没有被上传
	if _, ok := mockClient.files["/root/cloudclaw/.env"]; ok {
		t.Error("failed file should not be uploaded")
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
