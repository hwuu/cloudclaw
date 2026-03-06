package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hwuu/cloudclaw/internal/config"
)

// TestCredentials_SaveAndLoad 测试凭证的保存和加载
func TestCredentials_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "credentials")

	cred := &config.Credentials{
		AccessKeyID:     "test_key_id",
		AccessKeySecret: "test_key_secret",
		Region:          "ap-southeast-1",
	}

	// 保存
	err := config.SaveCredentialsTo(path, cred)
	if err != nil {
		t.Fatalf("SaveCredentialsTo() error = %v", err)
	}

	// 加载
	loaded, err := config.LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("LoadCredentialsFrom() error = %v", err)
	}

	if loaded.AccessKeyID != cred.AccessKeyID {
		t.Errorf("AccessKeyID = %s, want %s", loaded.AccessKeyID, cred.AccessKeyID)
	}
	if loaded.AccessKeySecret != cred.AccessKeySecret {
		t.Errorf("AccessKeySecret = %s, want %s", loaded.AccessKeySecret, cred.AccessKeySecret)
	}
	if loaded.Region != cred.Region {
		t.Errorf("Region = %s, want %s", loaded.Region, cred.Region)
	}
}

// TestCredentials_FileNotFound 测试凭证文件不存在
func TestCredentials_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "credentials")

	_, err := config.LoadCredentialsFrom(path)
	if err == nil {
		t.Fatal("LoadCredentialsFrom() error = nil, want error")
	}
}

// TestCredentials_MissingKey 测试缺少 AccessKeyID
func TestCredentials_MissingKey(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "credentials")

	// 只写 secret，不写 id
	content := "access_key_secret=secret\n"
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.LoadCredentialsFrom(path)
	if err == nil {
		t.Fatal("LoadCredentialsFrom() error = nil, want error for missing access_key_id")
	}
}

// TestCredentials_MissingSecret 测试缺少 AccessKeySecret
func TestCredentials_MissingSecret(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "credentials")

	// 只写 id，不写 secret
	content := "access_key_id=id\n"
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.LoadCredentialsFrom(path)
	if err == nil {
		t.Fatal("LoadCredentialsFrom() error = nil, want error for missing access_key_secret")
	}
}

// TestCredentials_CommentLines 测试注释行被忽略
func TestCredentials_CommentLines(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "credentials")

	content := `# 这是注释
access_key_id=test_id
# 这也是注释
access_key_secret=test_secret
region=ap-southeast-1
`
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, err := config.LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("LoadCredentialsFrom() error = %v", err)
	}

	if loaded.AccessKeyID != "test_id" {
		t.Errorf("AccessKeyID = %s, want test_id", loaded.AccessKeyID)
	}
}

// TestCredentials_FilePermission 测试文件权限
func TestCredentials_FilePermission(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "credentials")

	cred := &config.Credentials{
		AccessKeyID:     "test_id",
		AccessKeySecret: "test_secret",
		Region:          "ap-southeast-1",
	}

	err := config.SaveCredentialsTo(path, cred)
	if err != nil {
		t.Fatalf("SaveCredentialsTo() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// 检查权限是否为 0600
	if info.Mode().Perm() != 0600 {
		t.Errorf("File permission = %o, want 0600", info.Mode().Perm())
	}
}

// TestCredentials_EmptyLines 测试空行被忽略
func TestCredentials_EmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "credentials")

	content := `

access_key_id=test_id

access_key_secret=test_secret

region=ap-southeast-1

`
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	loaded, err := config.LoadCredentialsFrom(path)
	if err != nil {
		t.Fatalf("LoadCredentialsFrom() error = %v", err)
	}

	if loaded.AccessKeyID != "test_id" {
		t.Errorf("AccessKeyID = %s, want test_id", loaded.AccessKeyID)
	}
}
