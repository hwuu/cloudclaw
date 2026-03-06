package config

import (
	"testing"

	"github.com/hwuu/cloudclaw/internal/config"
)

// TestState_HasVPC 测试 VPC 资源判断
func TestState_HasVPC(t *testing.T) {
	s := &config.State{
		Resources: config.Resources{
			VPC: config.VPCResource{ID: "vpc-123"},
		},
	}

	if !s.HasVPC() {
		t.Error("HasVPC() = false, want true")
	}

	s.Resources.VPC.ID = ""
	if s.HasVPC() {
		t.Error("HasVPC() = true, want false")
	}
}

// TestState_HasVSwitch 测试 VSwitch 资源判断
func TestState_HasVSwitch(t *testing.T) {
	s := &config.State{
		Resources: config.Resources{
			VSwitch: config.VSwitchResource{ID: "vsw-123"},
		},
	}

	if !s.HasVSwitch() {
		t.Error("HasVSwitch() = false, want true")
	}

	s.Resources.VSwitch.ID = ""
	if s.HasVSwitch() {
		t.Error("HasVSwitch() = true, want false")
	}
}

// TestState_HasSecurityGroup 测试安全组资源判断
func TestState_HasSecurityGroup(t *testing.T) {
	s := &config.State{
		Resources: config.Resources{
			SecurityGroup: config.SecurityGroupResource{ID: "sg-123"},
		},
	}

	if !s.HasSecurityGroup() {
		t.Error("HasSecurityGroup() = false, want true")
	}

	s.Resources.SecurityGroup.ID = ""
	if s.HasSecurityGroup() {
		t.Error("HasSecurityGroup() = true, want false")
	}
}

// TestState_HasECS 测试 ECS 资源判断
func TestState_HasECS(t *testing.T) {
	s := &config.State{
		Resources: config.Resources{
			ECS: config.ECSResource{ID: "i-123"},
		},
	}

	if !s.HasECS() {
		t.Error("HasECS() = false, want true")
	}

	s.Resources.ECS.ID = ""
	if s.HasECS() {
		t.Error("HasECS() = true, want false")
	}
}

// TestState_HasEIP 测试 EIP 资源判断
func TestState_HasEIP(t *testing.T) {
	s := &config.State{
		Resources: config.Resources{
			EIP: config.EIPResource{ID: "eip-123"},
		},
	}

	if !s.HasEIP() {
		t.Error("HasEIP() = false, want true")
	}

	s.Resources.EIP.ID = ""
	if s.HasEIP() {
		t.Error("HasEIP() = true, want false")
	}
}

// TestState_HasSSHKeyPair 测试 SSH 密钥对资源判断
func TestState_HasSSHKeyPair(t *testing.T) {
	s := &config.State{
		Resources: config.Resources{
			SSHKeyPair: config.SSHKeyPairResource{Name: "cloudclaw-ssh-key"},
		},
	}

	if !s.HasSSHKeyPair() {
		t.Error("HasSSHKeyPair() = false, want true")
	}

	s.Resources.SSHKeyPair.Name = ""
	if s.HasSSHKeyPair() {
		t.Error("HasSSHKeyPair() = true, want false")
	}
}

// TestState_IsComplete 测试资源完整性判断（表驱动测试）
func TestState_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*config.State)
		expected bool
	}{
		{
			name:     "all_resources_present",
			setup:    func(s *config.State) {},
			expected: true,
		},
		{
			name: "missing_vpc",
			setup: func(s *config.State) {
				s.Resources.VPC.ID = ""
			},
			expected: false,
		},
		{
			name: "missing_vswitch",
			setup: func(s *config.State) {
				s.Resources.VSwitch.ID = ""
			},
			expected: false,
		},
		{
			name: "missing_security_group",
			setup: func(s *config.State) {
				s.Resources.SecurityGroup.ID = ""
			},
			expected: false,
		},
		{
			name: "missing_ecs",
			setup: func(s *config.State) {
				s.Resources.ECS.ID = ""
			},
			expected: false,
		},
		{
			name: "missing_eip",
			setup: func(s *config.State) {
				s.Resources.EIP.ID = ""
			},
			expected: false,
		},
		{
			name: "missing_ssh_key_pair",
			setup: func(s *config.State) {
				s.Resources.SSHKeyPair.Name = ""
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &config.State{
				Resources: config.Resources{
					VPC:           config.VPCResource{ID: "vpc-123"},
					VSwitch:       config.VSwitchResource{ID: "vsw-123"},
					SecurityGroup: config.SecurityGroupResource{ID: "sg-123"},
					ECS:           config.ECSResource{ID: "i-123"},
					EIP:           config.EIPResource{ID: "eip-123"},
					SSHKeyPair:    config.SSHKeyPairResource{Name: "cloudclaw-ssh-key"},
				},
			}
			tt.setup(s)
			if got := s.IsComplete(); got != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestNewState 测试新状态创建
func TestNewState(t *testing.T) {
	s := config.NewState("ap-southeast-1", "ubuntu_24_04")

	if s.Version != config.StateFileVersion {
		t.Errorf("Version = %s, want %s", s.Version, config.StateFileVersion)
	}
	if s.Region != "ap-southeast-1" {
		t.Errorf("Region = %s, want ap-southeast-1", s.Region)
	}
	if s.OSImage != "ubuntu_24_04" {
		t.Errorf("OSImage = %s, want ubuntu_24_04", s.OSImage)
	}
	if s.HasVPC() {
		t.Error("NewState() should not have VPC")
	}
}
