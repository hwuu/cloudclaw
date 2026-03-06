package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hwuu/cloudclaw/internal/config"
)

// TestBackup_SaveAndLoad 测试备份文件的保存和加载
func TestBackup_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	backup := &config.Backup{
		CloudClawVersion: "v0.7.0",
		SnapshotID:       "s-test123",
		CreatedAt:        "2026-03-06T00:00:00Z",
		Region:           "ap-southeast-1",
		DiskSize:         60,
		Domain:           "example.com",
	}

	// 保存
	err := config.SaveBackupTo(tempDir, backup)
	if err != nil {
		t.Fatalf("SaveBackupTo() error = %v", err)
	}

	// 加载
	loaded, err := config.LoadBackupFrom(tempDir)
	if err != nil {
		t.Fatalf("LoadBackupFrom() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadBackupFrom() returned nil")
	}
	if loaded.CloudClawVersion != backup.CloudClawVersion {
		t.Errorf("CloudClawVersion = %s, want %s", loaded.CloudClawVersion, backup.CloudClawVersion)
	}
	if loaded.SnapshotID != backup.SnapshotID {
		t.Errorf("SnapshotID = %s, want %s", loaded.SnapshotID, backup.SnapshotID)
	}
	if loaded.Domain != backup.Domain {
		t.Errorf("Domain = %s, want %s", loaded.Domain, backup.Domain)
	}
}

// TestBackup_LoadNotFound 测试备份文件不存在的情况
func TestBackup_LoadNotFound(t *testing.T) {
	tempDir := t.TempDir()

	backup, err := config.LoadBackupFrom(tempDir)
	if err != nil {
		t.Fatalf("LoadBackupFrom() error = %v, want nil", err)
	}
	if backup != nil {
		t.Errorf("LoadBackupFrom() = %v, want nil", backup)
	}
}

// TestBackup_Delete 测试删除备份文件
func TestBackup_Delete(t *testing.T) {
	tempDir := t.TempDir()

	backup := &config.Backup{SnapshotID: "s-test123"}
	err := config.SaveBackupTo(tempDir, backup)
	if err != nil {
		t.Fatalf("SaveBackupTo() error = %v", err)
	}

	// 删除
	err = config.DeleteBackupFrom(tempDir)
	if err != nil {
		t.Fatalf("DeleteBackupFrom() error = %v", err)
	}

	// 验证已删除
	loaded, err := config.LoadBackupFrom(tempDir)
	if err != nil {
		t.Fatalf("LoadBackupFrom() error = %v", err)
	}
	if loaded != nil {
		t.Error("DeleteBackupFrom() did not delete the file")
	}
}

// TestBackup_JSONFormat 测试备份文件 JSON 格式
func TestBackup_JSONFormat(t *testing.T) {
	tempDir := t.TempDir()

	backup := &config.Backup{
		CloudClawVersion: "v0.7.0",
		SnapshotID:       "s-test123",
		CreatedAt:        "2026-03-06T00:00:00Z",
		Region:           "ap-southeast-1",
		DiskSize:         60,
		Domain:           "example.com",
	}

	err := config.SaveBackupTo(tempDir, backup)
	if err != nil {
		t.Fatalf("SaveBackupTo() error = %v", err)
	}

	// 读取原始文件内容
	path := filepath.Join(tempDir, config.BackupFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// 验证 JSON 格式（应该包含缩进）
	expectedStart := "{\n"
	if string(data)[:len(expectedStart)] != expectedStart {
		t.Errorf("JSON should be indented, got: %s...", string(data)[:20])
	}
}

// TestBackup_CorruptedJSON 测试损坏的 JSON 文件
func TestBackup_CorruptedJSON(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, config.BackupFileName)

	// 写入无效的 JSON
	invalidJSON := `{invalid json content}`
	err := os.WriteFile(path, []byte(invalidJSON), 0600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.LoadBackupFrom(tempDir)
	if err == nil {
		t.Fatal("LoadBackupFrom() error = nil, want JSON parse error")
	}
}

// TestBackup_EmptyFile 测试空文件
func TestBackup_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, config.BackupFileName)

	// 写入空内容
	err := os.WriteFile(path, []byte(""), 0600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = config.LoadBackupFrom(tempDir)
	if err == nil {
		t.Fatal("LoadBackupFrom() error = nil, want error for empty file")
	}
}
