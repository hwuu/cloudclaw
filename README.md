# CloudClaw

**一键部署 OpenClaw 到阿里云 ECS**

CloudClaw 是一个类似 CloudCode 的 CLI 工具，用于一键部署 OpenClaw 个人 AI 助手到阿里云 ECS。

## 功能特性

- **一键部署**: `cloudclaw deploy` 完成所有云资源创建和应用部署
- **HTTPS 自动配置**: 使用 Caddy 自动管理 HTTPS 证书
- **Gateway Token 认证**: 内置 Gateway 访问令牌认证
- **多频道支持**: 支持飞书、Telegram、Discord、Slack 等聊天渠道
- **停机省钱**: suspend/resume 功能，停机仅收磁盘费
- **快照恢复**: destroy 时可选保留快照
- **OSS 状态存储**: 跨机器共享状态，支持中断恢复
- **分布式锁**: 防止并发操作冲突

## 快速开始

### 安装

```bash
# 从源码构建
git clone https://github.com/hwuu/cloudclaw.git
cd cloudclaw
make build
```

### 使用

```bash
# 1. 初始化配置（输入阿里云凭证）
./bin/cloudclaw init

# 2. 一键部署 OpenClaw
./bin/cloudclaw deploy

# 3. 查看部署状态
./bin/cloudclaw status

# 4. 停机省钱（停止计费，仅收磁盘费）
./bin/cloudclaw suspend

# 5. 恢复运行
./bin/cloudclaw resume

# 6. 销毁所有资源
./bin/cloudclaw destroy
```

### 可用命令

| 命令 | 说明 |
|------|------|
| `init` | 配置阿里云凭证并创建 OSS bucket |
| `deploy` | 一键部署 OpenClaw |
| `status` | 查看部署状态 |
| `destroy` | 销毁所有云资源（可选保留快照） |
| `suspend` | 停机（停止计费，仅收磁盘费） |
| `resume` | 恢复运行 |
| `logs` | 查看容器日志 |
| `ssh` | SSH 登录 ECS |
| `exec` | 在容器内执行命令 |
| `plugins` | 管理 OpenClaw 插件 |
| `channels` | 管理聊天渠道 |
| `config` | 查看或修改配置 |
| `version` | 显示版本信息 |

## 支持的聊天渠道

| 渠道 | 配置难度 | 推荐度 |
|------|----------|--------|
| Telegram | ⭐⭐ 简单 | ⭐⭐⭐⭐⭐ 最推荐 |
| 飞书 (Feishu) | ⭐⭐⭐ 中等 | ⭐⭐⭐⭐ 推荐 |
| Discord | ⭐⭐ 简单 | ⭐⭐⭐⭐ 推荐 |
| Slack | ⭐⭐⭐ 中等 | ⭐⭐⭐⭐ 推荐 |
| WebChat | ⭐ 最简单 | ⭐⭐⭐ 通用 |

## 架构

```
客户端 (聊天应用)
       │
       ▼
    Internet
       │
       ▼
    [EIP] ─────► ECS 实例
                      │
                      ├─── [Caddy 容器] (80/443)
                      │         │
                      │         ▼
                      │    [openclaw-gateway 容器] (18789)
                      │
                      └─── Docker Daemon
```

## 成本

| 资源 | 月费用 | 备注 |
|------|--------|------|
| ECS | ~$20/月 | ecs.e-c1m2.large，按量付费 |
| EIP | ~$2/月 | 按量付费 |
| OSS | ~$0.01/月 | 状态存储 |
| 停机状态 | ~$3/月 | 仅磁盘费 |

## 开发

```bash
# 构建
make build

# 运行测试
make test

# 清理构建产物
make clean
```

## 参考资料

- [设计文档](docs/v0.1/design.md)
- [实现计划](docs/v0.1/implementation-plan.md)
- [CloudCode](https://github.com/hwuu/cloudcode) - 本项目的基础
- [OpenClaw](https://github.com/openclaw/openclaw) - 个人 AI 助手

## License

MIT
