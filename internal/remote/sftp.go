package remote

// sftp.go 封装 SFTP 文件上传操作，用于将渲染后的配置文件上传到 ECS 实例。

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// sshClient SSH 客户端实现
type sshClient struct {
	client *ssh.Client
}

// NewSSHDialFunc 创建 SSH 连接的 DialFunc
func NewSSHDialFunc(host string, port int, user string, privateKey []byte) DialFunc {
	return func() (SSHClient, error) {
		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("解析 SSH 私钥失败：%w", err)
		}

		config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		client, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			return nil, fmt.Errorf("SSH 连接失败 (%s): %w", addr, err)
		}

		return &sshClient{client: client}, nil
	}
}

// RunCommand 在远程执行命令，支持 context 超时取消。
// 返回 stdout 内容；失败时错误信息包含 stderr。
func (c *sshClient) RunCommand(ctx context.Context, cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH session 失败：%w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- session.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			return stdout.String(), fmt.Errorf("命令执行失败：%w\nstderr: %s", err, stderr.String())
		}
		return stdout.String(), nil
	}
}

func (c *sshClient) Close() error {
	return c.client.Close()
}

// sftpClient SFTP 客户端实现
type sftpClient struct {
	sftpClient *sftp.Client
	sshClient  *ssh.Client
}

// NewSFTPClient 创建 SFTP 客户端
func NewSFTPClient(host string, port int, user string, privateKey []byte) (SFTPClient, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("解析 SSH 私钥失败：%w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	sshConn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败：%w", err)
	}

	sftpConn, err := sftp.NewClient(sshConn)
	if err != nil {
		sshConn.Close()
		return nil, fmt.Errorf("SFTP 连接失败：%w", err)
	}

	return &sftpClient{
		sftpClient: sftpConn,
		sshClient:  sshConn,
	}, nil
}

// UploadFile 上传文件内容到远程路径（自动创建父目录）
func (c *sftpClient) UploadFile(localContent []byte, remotePath string) error {
	// 自动创建远程目录
	dir := filepath.Dir(remotePath)
	if err := c.sftpClient.MkdirAll(dir); err != nil {
		return fmt.Errorf("创建远程目录 %s 失败：%w", dir, err)
	}

	f, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("创建远程文件 %s 失败：%w", remotePath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, bytes.NewReader(localContent)); err != nil {
		return fmt.Errorf("写入远程文件 %s 失败：%w", remotePath, err)
	}

	return nil
}

func (c *sftpClient) Close() error {
	c.sftpClient.Close()
	return c.sshClient.Close()
}

// SFTPClient 抽象 SFTP 文件上传，支持 mock 测试
type SFTPClient interface {
	UploadFile(localContent []byte, remotePath string) error
	Close() error
}

// UploadFiles 批量上传文件（remotePath → content 映射），任一失败立即返回错误
func UploadFiles(client SFTPClient, files map[string][]byte) error {
	for remotePath, content := range files {
		if err := client.UploadFile(content, remotePath); err != nil {
			return fmt.Errorf("failed to upload %s: %w", remotePath, err)
		}
	}
	return nil
}

// GetPublicIP 通过外部服务（ipify）获取用户公网 IP，用于限制 SSH 安全组规则。
// 支持通过 CLOUDCLAW_PUBLIC_IP 环境变量覆盖（测试/代理场景）。
func GetPublicIP() (string, error) {
	if ip := os.Getenv("CLOUDCLAW_PUBLIC_IP"); ip != "" {
		return ip, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "", fmt.Errorf("无法获取公网 IP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取公网 IP 响应失败：%w", err)
	}

	return strings.TrimSpace(string(body)), nil
}
