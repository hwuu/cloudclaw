package template

import (
	"strings"
	"testing"

	tmpl "github.com/hwuu/cloudclaw/internal/template"
)

// TestRenderTemplate_Caddyfile 测试 Caddyfile 模板渲染
func TestRenderTemplate_Caddyfile(t *testing.T) {
	data := &tmpl.TemplateData{
		Domain: "example.com",
	}

	content, err := tmpl.RenderTemplate("templates/Caddyfile.tmpl", data)
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}

	rendered := string(content)
	if !strings.Contains(rendered, "example.com") {
		t.Errorf("Rendered Caddyfile should contain 'example.com', got: %s", rendered[:50])
	}
}

// TestRenderTemplate_Env 测试 .env 模板渲染
func TestRenderTemplate_Env(t *testing.T) {
	data := &tmpl.TemplateData{
		Domain:       "example.com",
		GatewayToken: "test_token_123",
	}

	content, err := tmpl.RenderTemplate("templates/env.tmpl", data)
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}

	rendered := string(content)
	if !strings.Contains(rendered, "test_token_123") {
		t.Errorf("Rendered .env should contain token, got: %s", rendered)
	}
}

// TestRenderTemplate_DockerCompose 测试 docker-compose.yml 模板渲染
func TestRenderTemplate_DockerCompose(t *testing.T) {
	data := &tmpl.TemplateData{
		Domain:  "example.com",
		Version: "v0.7.0",
	}

	content, err := tmpl.RenderTemplate("templates/docker-compose.yml.tmpl", data)
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}

	rendered := string(content)
	if !strings.Contains(rendered, "v0.7.0") {
		t.Errorf("Rendered docker-compose.yml should contain version, got: %s", rendered[:100])
	}
}

// TestRenderTemplate_VersionDefault 测试 Version 默认值
func TestRenderTemplate_VersionDefault(t *testing.T) {
	data := &tmpl.TemplateData{
		Domain:       "example.com",
		GatewayToken: "test_token",
		// Version 为空
	}

	_, err := tmpl.RenderAll(data)
	if err != nil {
		t.Fatalf("RenderAll() error = %v", err)
	}
	// RenderAll 内部会将空 Version 设置为 "latest"
}

// TestRenderAll 测试渲染所有文件
func TestRenderAll(t *testing.T) {
	data := &tmpl.TemplateData{
		Domain:       "test.example.com",
		GatewayToken: "token_xyz",
		Version:      "v1.0.0",
	}

	files, err := tmpl.RenderAll(data)
	if err != nil {
		t.Fatalf("RenderAll() error = %v", err)
	}

	expectedPaths := []string{
		"~/cloudclaw/Caddyfile",
		"~/cloudclaw/.env",
		"~/cloudclaw/docker-compose.yml",
	}

	for _, path := range expectedPaths {
		if _, ok := files[path]; !ok {
			t.Errorf("RenderAll() missing file: %s", path)
		}
	}

	// 验证 .env 内容包含 token
	envContent := string(files["~/cloudclaw/.env"])
	if !strings.Contains(envContent, "token_xyz") {
		t.Errorf(".env should contain token")
	}
}

// TestRenderTemplate_NotFound 测试模板不存在
func TestRenderTemplate_NotFound(t *testing.T) {
	data := &tmpl.TemplateData{
		GatewayToken: "test",
	}

	_, err := tmpl.RenderTemplate("templates/nonexistent.tmpl", data)
	if err == nil {
		t.Fatal("RenderTemplate() error = nil, want error for nonexistent template")
	}
}

// TestTemplateFileList 测试模板文件列表
func TestTemplateFileList(t *testing.T) {
	files := tmpl.TemplateFileList()
	if len(files) != 3 {
		t.Errorf("TemplateFileList() returned %d files, want 3", len(files))
	}
}

// TestStaticFileList 测试静态文件列表
func TestStaticFileList(t *testing.T) {
	files := tmpl.StaticFileList()
	if len(files) != 0 {
		t.Errorf("StaticFileList() returned %d files, want 0", len(files))
	}
}
