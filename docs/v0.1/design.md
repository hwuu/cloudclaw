# CloudClaw v0.1.0 设计文档

## 目录

- [1. 背景与目标](#1-背景与目标)
  - [1.1 OpenClaw 介绍](#11-openclaw-介绍)
  - [1.2 v0.1.0 目标](#12-v010-目标)
  - [1.3 非目标](#13-非目标)
- [2. 设计决策](#2-设计决策)
  - [2.1 架构复用：为什么直接复用 CloudCode 而非从零开始](#21-架构复用为什么直接复用-cloudcode-而非从零开始)
  - [2.2 认证方案：为什么用 Gateway Token 而非 Authelia](#22-认证方案为什么用-gateway-token-而非-authelia)
  - [2.3 状态存储方案决策](#23-状态存储方案决策)
- [3. 组件设计](#3-组件设计)
  - [3.1 部署架构](#31-部署架构)
  - [3.2 状态管理](#32-状态管理)
  - [3.3 配置管理](#33-配置管理)
  - [3.4 应用部署](#34-应用部署)
  - [3.5 并发控制](#35-并发控制)
  - [3.6 幂等性保证](#36-幂等性保证)
  - [3.7 Edge Case 处理](#37-edge-case-处理)
- [4. 接口设计](#4-接口设计)
  - [4.1 CLI 命令](#41-cli-命令)
  - [4.2 与 CloudCode 的差异](#42-与-cloudcode-的差异)
- [5. 实现规划](#5-实现规划)
  - [5.1 文件清单](#51-文件清单)
  - [5.2 实现步骤](#52-实现步骤)
- [6. 成本影响](#6-成本影响)
- [7. 测试要点](#7-测试要点)
- [变更记录](#变更记录)

---

## 1. 背景与目标

### 1.1 OpenClaw 介绍

OpenClaw (https://github.com/openclaw/openclaw) 是一个个人 AI 助手，支持多频道接入：

- **核心服务**: Gateway (WebSocket 服务器，端口 18789)
- **支持频道**: WhatsApp, Telegram, Slack, Discord, Signal, iMessage 等
- **部署方式**: Docker Compose
- **认证方式**: Gateway Token (内置)
- **技术栈**: Node.js + TypeScript

**与 OpenCode 的对比**:

| 特性 | OpenCode | OpenClaw |
|------|----------|----------|
| 定位 | AI 编程助手 | 个人 AI 助手 (多频道) |
| 入口 | Web UI | Gateway WS + 多频道客户端 |
| 端口 | 4096 (HTTP) | 18789 (WS) |
| 认证 | Authelia + Passkey | Gateway Token |

### 1.2 v0.1.0 目标

CloudClaw 是一个类似 CloudCode 的 CLI 工具，用于一键部署 OpenClaw 到阿里云 ECS。

| 目标 | 说明 |
|------|------|
| **一键部署** | `cloudclaw deploy` 完成所有云资源创建和应用部署 |
| **HTTPS 自动配置** | 使用 Caddy 自动管理 HTTPS 证书 |
| **Gateway Token 认证** | 内置 Gateway 访问令牌认证 |
| **多频道配置** | 支持配置各种频道 (Telegram, Discord, Slack 等) 的凭证 |
| **停机省钱** | suspend/resume 功能，停机仅收磁盘费 |
| **快照恢复** | destroy 时可选保留快照 |
| **OSS 状态存储** | 跨机器共享状态，支持中断恢复 |
| **分布式锁** | 防止并发操作冲突 |
| **运维命令** | logs/ssh/exec 等日常运维支持 |

### 1.3 非目标

- **不支持和平方**：专注于单用户场景
- **不做多云支持**：首期仅支持阿里云
- **不做 Web Terminal**：OpenClaw 自带 CLI，无需 ttyd
- **不做复杂认证**：使用 Gateway Token 替代 Authelia

---

## 2. 设计决策

### 2.1 架构复用：为什么直接复用 CloudCode 而非从零开始

#### 2.1.1 方案对比

| 方案 | 优点 | 缺点 |
|------|------|------|
| **直接复用 CloudCode** | 开发成本低，代码质量有保障，架构经过验证 | 需要清理不相关的功能 (Authelia/ttyd) |
| **从零开始** | 完全定制，无历史包袱 | 开发周期长，容易引入新 bug |
| **Fork CloudCode** | 保留完整历史，便于对比 | 代码耦合度高，难以剥离 |

**决策**：复用 CloudCode 架构和代码，新建项目。

#### 2.1.2 复用策略

| 模块 | 复用程度 | 说明 |
|------|----------|------|
| `internal/alicloud/` | 100% 复用 | 阿里云 SDK 封装完全相同 |
| `internal/config/` | 90% 复用 | 配置结构需调整 (去掉 Authelia 相关) |
| `internal/deploy/` | 80% 复用 | 部署流程相同，应用层不同 |
| `internal/template/` | 50% 复用 | 模板内容不同，框架相同 |
| `internal/remote/` | 100% 复用 | SSH/SFTP 完全相同 |
| `cmd/` | 70% 复用 | 命令名称、帮助文本需修改 |

### 2.2 认证方案：为什么用 Gateway Token 而非 Authelia

#### 2.2.1 方案对比

| 方案 | 优点 | 缺点 |
|------|------|------|
| **Authelia (CloudCode 方案)** | 功能完整，支持多因素认证 | 配置复杂，资源占用高，需额外容器 |
| **Gateway Token (OpenClaw 原生)** | 简单，无需额外容器，OpenClaw 内置支持 | 功能单一，仅 Token 认证 |
| **Nginx + Basic Auth** | 简单 | 不安全，密码明文传输，不推荐 |

**决策**：使用 OpenClaw 原生的 Gateway Token 认证。

#### 2.2.2 选择 Gateway Token 的理由

1. **OpenClaw 内置支持**：Gateway 原生支持 Token 认证，无需额外配置
2. **简化架构**：无需 Authelia 容器，减少资源占用和复杂度
3. **配置简单**：只需设置 `OPENCLAW_GATEWAY_TOKEN` 环境变量
4. **安全性足够**：Token 随机生成 (建议 32 字节)，长度足够，爆破难度大
5. **与 CloudCode 差异化**：CloudClaw 定位更轻量，面向单用户

#### 2.2.3 Caddy 配置方案

使用 Caddy 的 `reverse_proxy` 实现 WebSocket 和普通请求的反向代理：

```Caddyfile
openclaw.example.com {
    reverse_proxy openclaw-gateway:18789 {
        header_up Upgrade {http.request.header.Upgrade}
        header_up Connection {http.request.header.Connection}
    }

    encode gzip

    log {
        output file /var/log/caddy/access.log
    }
}
```

---

### 2.3 状态存储方案决策

#### 2.3.1 方案对比

| 方案 | 优点 | 缺点 |
|------|------|------|
| **本地文件** | 简单，无网络依赖 | 无法跨机器，中断无法恢复，无锁机制 |
| **ECS 标签** | 无额外资源 | deploy 初期无 ECS；destroy 后丢失状态 |
| **OSS bucket** | 全流程可用，支持条件写入作锁，跨机器共享，成本极低 | 需创建额外资源，网络请求有延迟 |
| **云数据库 (TableStore/RDS)** | 功能完整，强一致性 | 成本高，复杂度过大，增加依赖 |
| **DNS TXT 记录** | 复用现有域名 | 不适合存大量状态，修改慢，有 TTL 延迟 |

**决策**：使用 OSS bucket 存储状态。

#### 2.3.2 选择 OSS 的理由

1. **跨机器共享**：状态存储在云端，任意机器可访问
2. **分布式锁**：OSS 支持条件写入（`If-None-Match: *`），天然适合分布式锁
3. **全生命周期覆盖**：从 deploy 开始到 destroy 结束，状态始终可用
4. **成本极低**：OSS 标准存储 ~0.12 元/GB/月，state.json 几 KB，费用可忽略
5. **无需额外依赖**：用户已有阿里云账号，OSS 是同生态服务

#### 2.3.3 本地 vs OSS 职责划分

| 数据类型 | 本地文件 | OSS 文件 | 理由 |
|----------|----------|----------|------|
| AccessKey ID/Secret | ✓ `~/.cloudclaw/credentials` | ✗ | 敏感信息，不应上传云端 |
| SSH 私钥 | ✓ `~/.cloudclaw/ssh_key` | ✗ | 私钥，仅本地持有 |
| 部署状态 (state.json) | ✓ 本地缓存 | ✓ 主存储 | 跨机器共享、分布式锁 |
| 快照元数据 (backup.json) | ✗ | ✓ | 跨机器共享 |
| 操作历史 | ✗ | ✓ `history/` | 审计追踪 |
| 模板文件 | ✓ 嵌入二进制 + 本地覆盖 | ✗ | 允许用户自定义 |

---

## 3. 组件设计

### 3.1 部署架构

#### 3.1.1 架构图

```
客户端 (多频道/WS)
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
                      │              │
                      │              ▼
                      │         [openclaw-cli 容器]
                      │
                      └─── Docker Daemon
```

#### 3.1.2 云资源清单

| 资源 | 说明 | 配置 |
|------|------|------|
| VPC | 私有网络 | 192.168.0.0/16 |
| VSwitch | 交换机 | 192.168.1.0/24, Zone A |
| 安全组 | 访问控制 | 入站：80/443/22 |
| ECS | 计算实例 | ecs.e-c1m2.large (2C4G) |
| EIP | 弹性公网 IP | 按量付费 |
| OSS | 状态存储 | cloudclaw-state-<account_id> |
| SSH 密钥对 | SSH 登录 | cloudclaw-ssh-key |

#### 3.1.3 网络规则

| 端口 | 协议 | 源 | 目的 | 说明 |
|------|------|----|----|------|
| 22 | TCP | 用户配置/0.0.0.0 | ECS | SSH |
| 80 | TCP | 0.0.0.0/0 | ECS | HTTP (Caddy 自动 HTTPS) |
| 443 | TCP | 0.0.0.0/0 | ECS | HTTPS |
| 18789 | TCP | 内部 | Gateway | 内部服务，不暴露 |

### 3.2 状态管理

#### 3.2.1 状态定义

完全复用 CloudCode v0.3 的状态机设计：

| status | 含义 | 可转换到 |
|--------|------|----------|
| `deploying` | 部署中 | `running`（成功），失败时保持 `deploying`（从断点恢复） |
| `running` | 运行中 | `suspending`, `destroying` |
| `suspending` | 停机中 | `suspended`（成功），失败时回退到 `running` |
| `suspended` | 已停机 | `resuming`, `destroying` |
| `resuming` | 恢复中 | `running`（成功），失败时回退到 `suspended` |
| `destroying` | 销毁中 | `destroyed`（成功），失败时回退到 `previous_status` |
| `destroyed` | 已销毁（有快照） | `deploying`（从快照恢复）；无快照时直接删除 state.json |

**`previous_status` 字段说明**：

| 场景 | previous_status 值 | 说明 |
|------|-------------------|------|
| deploy 开始 | 清空 | 全新部署或从快照恢复 |
| suspend 开始 | `running` | 记录进入 suspending 前的状态 |
| suspend 成功 | 保持 `running` | 用于 destroy 时回退 |
| resume 开始 | `suspended` | 记录进入 resuming 前的状态 |
| resume 成功 | 清空 | 恢复正常状态 |
| destroy 开始 | 当前状态 (`running`/`suspended`) | 用于 destroy 失败时回退 |
| destroy 成功 | 保持原值 | destroyed 状态保留 |
| destroy 失败 | 恢复到 previous_status | 回退到 destroy 前的状态 |

**State Transition Diagram**:

```
                                 [no state]
                                      |
                                 start deploy
                                      v
                              +---------------+
                              | [deploying]   | <---------------------+
                              +---------------+                       |
                                   |        |                         |
                            success|        |failed                   |
                                   v        |                         |
                          +---------------+ |                         |
                          |  [running]    | |                         |
                          +---------------+ |                         |
                               |       |    |                         |
                        suspend|  destroy   |                         |
                               v       |    |                         |
                      +---------------+ |   |                         |
                      | [suspending]  | |   |                         |
                      +---------------+ |   |                         |
                           |        |   |   |                         |
                    success|        |failed  |                        |
                           v        v   |   |                         |
                  +---------------+ +---+---+---+                     |
                  | [suspended]   | |  [running] |                    |
                  +---------------+ +------------+                    |
                       |       |                                      |
                 resume|  destroy                                     |
                       v       |                                      |
                  +---------------+                                   |
                  | [resuming]    |                                   |
                  +---------------+                                   |
                       |       |                                      |
                success|       |failed                                |
                       v       v                                      |
              +---------------+ +---------------+                     |
              |  [running]    | | [suspended]   |                     |
              +---------------+ +---------------+                     |
                                                                      |
  [running] or [suspended]                                            |
         |                                                            |
    destroy                                                           |
         v                                                            |
  +---------------+                                                   |
  | [destroying]  |                                                   |
  +---------------+                                                   |
       |        |                                                     |
  success|      |failed                                               |
       v        |                                                     |
  +---------------+                                                   |
  | [destroyed]   |                                                   |
  +---------------+                                                   |
         |                                                            |
   deploy from snapshot                                               |
         +------------------------------------------------------------+
```

#### 3.2.2 OSS 文件结构

```
cloudclaw-state-<account_id>/
├── state.json          # 当前状态
├── backup.json         # 快照元数据（可选）
├── .lock               # 分布式锁（临时）
└── history/            # 操作历史
    ├── 2026-03-05T10-00-00.json
    └── ...
```

**state.json** — 当前状态：

```json
{
  "version": "0.1.0",
  "status": "running",
  "previous_status": "",
  "region": "ap-southeast-1",
  "resources": {
    "vpc": {"id": "vpc-xxx", "cidr": "192.168.0.0/16"},
    "vswitch": {"id": "vsw-xxx", "zone_id": "ap-southeast-1a"},
    "security_group": {"id": "sg-xxx"},
    "ssh_key_pair": {"name": "cloudclaw-ssh-key"},
    "ecs": {"id": "i-xxx", "instance_type": "ecs.e-c1m2.large"},
    "eip": {"id": "eip-xxx", "ip": "47.100.1.1"}
  },
  "cloudclaw": {
    "domain": "ocl.example.com",
    "gateway_token": "xxx"
  },
  "updated_at": "2026-03-05T10:00:00Z"
}
```

**backup.json** — 快照元数据：

```json
{
  "snapshot_id": "s-t4nxxxxxxxxx",
  "cloudclaw_version": "0.1.0",
  "created_at": "2026-03-05T10:00:00Z",
  "region": "ap-southeast-1",
  "disk_size": 60
}
```

**.lock** — 分布式锁：

```json
{
  "operation": "suspend",
  "started_at": "2026-03-05T10:00:00Z",
  "client_id": "laptop-hwuu"
}
```

**history/<timestamp>.json** — 操作历史：

```json
{
  "timestamp": "2026-03-05T10:00:00Z",
  "operation": "suspend",
  "from_status": "running",
  "to_status": "suspended",
  "client_id": "laptop-hwuu",
  "success": true,
  "error": null,
  "duration_ms": 5000
}
```

#### 3.2.3 分布式锁实现

完全复用 CloudCode 的实现：

```go
type Lock struct {
    Operation string    `json:"operation"`
    StartedAt time.Time `json:"started_at"`
    ClientID  string    `json:"client_id"`  // hostname + PID
}

// AcquireLock 获取锁（原子操作：不存在才写入）
func AcquireLock(ossClient OSSClient, lock Lock) error {
    body, _ := json.Marshal(lock)
    _, err := ossClient.PutObject(".lock", body,
        oss.IfNoneMatch("*"))
    if isConditionFailed(err) {
        return ErrLockConflict
    }
    return err
}

// ReleaseLock 释放锁（原子操作：用 ETag 条件删除）
func ReleaseLock(ossClient OSSClient, clientID string) error {
    lock, etag, err := GetLockWithETag(ossClient)
    if err != nil {
        return err
    }
    if lock == nil {
        return nil  // 锁已不存在
    }
    if lock.ClientID != clientID {
        return ErrNotOwner
    }
    return ossClient.DeleteObject(".lock", oss.IfMatch(etag))
}

// ForceAcquireLock 强制接管（先删后写）
func ForceAcquireLock(ossClient OSSClient, lock Lock) error {
    if err := ossClient.DeleteObject(".lock"); err != nil && !isNotFound(err) {
        return fmt.Errorf("删除旧锁失败：%w", err)
    }
    body, _ := json.Marshal(lock)
    _, err := ossClient.PutObject(".lock", body,
        oss.IfNoneMatch("*"))
    if isConditionFailed(err) {
        return ErrLockConflict
    }
    return err
}

// RenewLock 续期（每 5 分钟调用一次）
func RenewLock(ossClient OSSClient, clientID string) error {
    lock, etag, err := GetLockWithETag(ossClient)
    if err != nil || lock == nil || lock.ClientID != clientID {
        return ErrNotOwner
    }
    lock.StartedAt = time.Now()
    body, _ := json.Marshal(lock)
    _, err = ossClient.PutObject(".lock", body, oss.IfMatch(etag))
    if isConditionFailed(err) {
        return ErrNotOwner
    }
    return err
}
```

**锁超时机制**：
- 锁的 `started_at` 超过 15 分钟视为过期
- 续期间隔 5 分钟，3 次未续期即过期
- 过期锁可被其他客户端自动接管（通过 `ForceAcquireLock`）

#### 3.2.4 中断恢复

完全复用 CloudCode 的设计：

| 场景 | 处理 |
|------|------|
| deploy 到一半断电 | 重启后检测 `status: deploying`，从断点恢复 |
| destroy 中断 | 重启后检测 `status: destroying`，检查快照是否存在 |
| suspend 到一半断电 | 重启后检测 `status: suspending`，查询 ECS 实际状态修正 |
| resume 到一半断电 | 重启后检测 `status: resuming`，查询 ECS 实际状态修正 |

#### 3.2.5 操作流程详解

**deploy 流程（带锁）**：

```
1. 检查本地凭证 → 不存在则报错退出
2. 读取 OSS state.json → 获取当前状态
3. [断点检查] 状态检查：
   - 不存在 → 全新部署
   - destroyed → 从快照恢复
   - deploying → 断点恢复
   - 其他 → 报错提示
4. 获取分布式锁：
   ├─ 调用 AcquireLock() → 成功则继续
   ├─ 锁冲突 (ErrLockConflict) → 读取 .lock 文件
   │   ├─ 锁过期 (started_at > 15min) → 调用 ForceAcquireLock() 自动接管
   │   └─ 锁未过期 → 提示"操作进行中，使用 --force 强制接管"并退出
   └─ 其他错误 → 报错退出
5. [续期] 启动后台 goroutine，每 5 分钟调用 RenewLock()
6. 写 OSS state (status: deploying, previous_status: "")
7. [断点恢复] 检查已创建的资源 → 跳过已存在步骤
8. 创建 VPC → 写 OSS state [CHECKPOINT]
   └─ 若失败：报错退出（锁由续期 goroutine 管理，自然过期）
9. 创建 VSwitch → 写 OSS state [CHECKPOINT]
   └─ 若失败：报错退出
10. 创建安全组 → 写 OSS state [CHECKPOINT]
    └─ 若失败：报错退出
11. 创建 SSH 密钥对 → 写 OSS state [CHECKPOINT]
    └─ 若失败：报错退出
12. 创建 ECS → 写 OSS state [CHECKPOINT]
    └─ 若失败：报错退出
13. 分配 EIP → 写 OSS state [CHECKPOINT]
    └─ 若失败：报错退出
14. 绑定 EIP → 写 OSS state [CHECKPOINT]
    └─ 若失败：报错退出
15. 等待 ECS 启动 + SSH 健康检查
    └─ 若失败：报错退出
16. SSH 上传 docker-compose.yml
    └─ 若失败：报错退出
17. SSH 上传 Caddyfile
    └─ 若失败：报错退出
18. SSH 上传 .env
    └─ 若失败：报错退出
19. SSH 执行 docker compose up -d
    └─ 若失败：报错退出
20. 等待健康检查通过
    └─ 若失败：报错退出
21. 写 OSS state (status: running, previous_status: "")
22. 写 OSS history
23. [释放锁] 停止续期 goroutine，调用 ReleaseLock()
```

**suspend 流程（带锁）**：

```
1. 读取 OSS state → 检查 status == running，否则报错
2. 获取分布式锁：
   ├─ 调用 AcquireLock() → 成功则继续
   ├─ 锁冲突 → 读取 .lock 文件
   │   ├─ 锁过期 (started_at > 15min) → 调用 ForceAcquireLock() 自动接管
   │   └─ 锁未过期 → 提示"操作进行中，使用 --force 强制接管"并退出
   └─ 其他错误 → 报错退出
3. [续期] 启动后台 goroutine，每 5 分钟调用 RenewLock()
4. 写 OSS state (status: suspending, previous_status: running)
5. SSH 执行 docker compose down (可选，直接停机也可)
   └─ 若失败：跳转到步骤 10，返回错误
6. 调用阿里云 StopInstance(StopChargingMode: true)
   └─ 若失败：跳转到步骤 10，返回错误
7. 等待实例状态变为 Stopped
   └─ 若失败：跳转到步骤 10，返回错误
8. 写 OSS state (status: suspended, previous_status: running)
9. 写 OSS history
10. [释放锁] 停止续期 goroutine，调用 ReleaseLock()
```

**resume 流程（带锁）**：

```
1. 读取 OSS state → 检查 status == suspended，否则报错
2. 获取分布式锁：
   ├─ 调用 AcquireLock() → 成功则继续
   ├─ 锁冲突 → 读取 .lock 文件
   │   ├─ 锁过期 (started_at > 15min) → 调用 ForceAcquireLock() 自动接管
   │   └─ 锁未过期 → 提示"操作进行中，使用 --force 强制接管"并退出
   └─ 其他错误 → 报错退出
3. [续期] 启动后台 goroutine，每 5 分钟调用 RenewLock()
4. 写 OSS state (status: resuming, previous_status: suspended)
5. 调用阿里云 StartInstance
   └─ 若失败：跳转到步骤 11，返回错误
6. 等待实例状态变为 Running
   └─ 若失败：跳转到步骤 11，返回错误
7. SSH 健康检查（尝试连接 22 端口）
   └─ 若失败：跳转到步骤 11，返回错误
8. SSH 执行 docker compose up -d
   └─ 若失败：跳转到步骤 11，返回错误
9. 写 OSS state (status: running, previous_status: "")
10. 写 OSS history
11. [释放锁] 停止续期 goroutine，调用 ReleaseLock()
```

**destroy 流程（带锁）**：

```
1. 读取 OSS state → 检查 status in [running, suspended]，否则报错
2. 提示是否保留快照 (y/N)
3. 获取分布式锁：
   ├─ 调用 AcquireLock() → 成功则继续
   ├─ 锁冲突 → 读取 .lock 文件
   │   ├─ 锁过期 (started_at > 15min) → 调用 ForceAcquireLock() 自动接管
   │   └─ 锁未过期 → 提示"操作进行中，使用 --force 强制接管"并退出
   └─ 其他错误 → 报错退出
4. [续期] 启动后台 goroutine，每 5 分钟调用 RenewLock()
5. 写 OSS state (status: destroying, previous_status: <当前状态>)
6. [可选] 创建快照：
   ├─ 调用阿里云 CreateSnapshot
   ├─ 等待快照完成
   └─ 写 OSS backup.json
   └─ 若失败：写 OSS state (status: previous_status)，跳转到步骤 10，返回错误
7. 删除资源（按依赖逆序）：
   ├─ 释放 EIP
   ├─ 删除 ECS
   ├─ 删除 SSH 密钥对
   ├─ 删除安全组
   ├─ 删除 VSwitch
   └─ 删除 VPC
   └─ 若失败：写 OSS state (status: previous_status)，跳转到步骤 10，返回错误
8. 写 OSS state：
   ├─ 保留快照：status: destroyed
   └─ 不保留快照：删除 state.json
9. 写 OSS history
10. [释放锁] 停止续期 goroutine，调用 ReleaseLock()
```

**图例说明**：
- `[CHECKPOINT]`：断点，部署中断后可从此步骤恢复
- `若失败：...`：失败分支处理，明确标注失败后的状态回退逻辑

### 3.3 配置管理

#### 3.3.1 交互配置项

`cloudclaw init` 和 `cloudclaw deploy` 时收集的参数：

| 配置项 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `AccessKeyID` | 是 | - | 阿里云 AccessKey ID |
| `AccessKeySecret` | 是 | - | 阿里云 AccessKey Secret |
| `Region` | 否 | `ap-southeast-1` | 阿里云区域 |
| `Domain` | 否 | `<EIP>.nip.io` | 域名（留空使用 nip.io） |
| `GatewayToken` | 是 | 随机生成 | OpenClaw Gateway Token |
| `SSHIPRestriction` | 否 | `0.0.0.0/0` | SSH 安全组源 IP 限制 |

**OpenClaw 频道配置**（可选，可后续通过 `cloudclaw config set` 设置）：

| 配置项 | 说明 |
|--------|------|
| `TelegramBotToken` | Telegram Bot Token |
| `DiscordBotToken` | Discord Bot Token |
| `SlackBotToken` | Slack Bot Token |
| `WhatsAppPhoneNumber` | WhatsApp 电话号码 |

#### 3.3.2 本地目录结构

```
~/.cloudclaw/
├── credentials     # AccessKeyID、AccessKeySecret、Region（权限 600）
└── ssh_key         # SSH 私钥（权限 600）
```

#### 3.3.3 环境变量模板

`.env` 文件模板：

```bash
# OpenClaw 配置
OPENCLAW_IMAGE=openclaw/openclaw:latest
OPENCLAW_CONFIG_DIR=/home/node/.openclaw
OPENCLAW_WORKSPACE_DIR=/home/node/.openclaw/workspace
OPENCLAW_GATEWAY_TOKEN=${GATEWAY_TOKEN}
OPENCLAW_GATEWAY_PORT=18789
OPENCLAW_BRIDGE_PORT=18790

# 频道配置（可选）
# TELEGRAM_BOT_TOKEN=xxx
# DISCORD_BOT_TOKEN=xxx
# SLACK_BOT_TOKEN=xxx
```

### 3.4 应用部署

#### 3.4.1 Docker Compose 模板

```yaml
services:
  caddy:
    image: caddy:2-alpine
    container_name: caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - openclaw-gateway
    networks:
      - cloudclaw-net

  openclaw-gateway:
    image: ${OPENCLAW_IMAGE:-openclaw/openclaw:latest}
    container_name: openclaw-gateway
    restart: unless-stopped
    environment:
      HOME: /home/node
      TERM: xterm-256color
      OPENCLAW_GATEWAY_TOKEN: ${OPENCLAW_GATEWAY_TOKEN}
    volumes:
      - ${OPENCLAW_CONFIG_DIR}:/home/node/.openclaw
      - ${OPENCLAW_WORKSPACE_DIR}:/home/node/.openclaw/workspace
    ports:
      - "${OPENCLAW_GATEWAY_PORT:-18789}:18789"
      - "${OPENCLAW_BRIDGE_PORT:-18790}:18790"
    init: true
    networks:
      - cloudclaw-net
    healthcheck:
      test: ["CMD", "node", "-e", "fetch('http://127.0.0.1:18789/healthz').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"]
      interval: 30s
      timeout: 5s
      retries: 5
      start_period: 20s

volumes:
  caddy_data:
  caddy_config:

networks:
  cloudclaw-net:
    driver: bridge
```

#### 3.4.2 Caddyfile 模板

```Caddyfile
{
    email ${ACME_EMAIL}
}

${DOMAIN} {
    reverse_proxy openclaw-gateway:18789 {
        header_up Upgrade {http.request.header.Upgrade}
        header_up Connection {http.request.header.Connection}
    }

    encode gzip

    log {
        output file /var/log/caddy/access.log
    }
}
```

**与 CloudCode 的差异**：
- CloudCode 有 Authelia 中间件，CloudClaw 没有
- CloudCode 有 WebSocket 路径特殊处理 (`/ws/*`)，CloudClaw 全部反向代理（Gateway 原生支持 WS）

#### 3.4.3 分阶段保存策略

| 阶段 | 保存内容 | 失败处理 |
|------|----------|----------|
| VPC 创建后 | vpc_id, cidr | 重试或回滚（删除 VPC） |
| VSwitch 创建后 | vswitch_id, zone_id | 重试或删除 VPC |
| 安全组创建后 | sg_id, rules | 重试或删除 VSwitch/VPC |
| ECS 创建后 | ecs_id, instance_type | 重试或删除安全组/VSwitch/VPC |
| EIP 分配后 | eip_id, eip_address | 重试或删除 ECS/... |
| EIP 绑定后 | eip_binding: true | 重试（资源保留） |
| 应用部署后 | app_deployed: true | 重试（资源保留） |

**分阶段保存策略说明**：
- 每个资源创建完成后立即写入 OSS state
- **先创建资源，再写 OSS**：若 OSS 写入失败但资源已创建，下次恢复时通过幂等性检查跳过已存在的资源
- 若 OSS 写入失败，操作中止并提示用户重试（已创建的资源不回滚，下次 deploy 从断点恢复）
- 阿里云 API 本身具备幂等性：重复创建同名 VPC/安全组等会返回已存在的资源 ID，不会重复创建

### 3.5 并发控制

#### 3.5.1 并发场景

| 场景 | 客户端 A | 客户端 B | 处理结果 |
|------|----------|----------|----------|
| 同时 deploy | 获取锁成功 | 获取锁失败 | B 提示"deploy 操作进行中，是否强制接管？[y/N]" |
| deploy 时执行 suspend | 正在 deploy | 尝试 suspend | B 获取锁失败，提示"部署中，请等待完成或强制接管" |
| suspend 时执行 resume | 正在 suspend | 尝试 resume | B 获取锁失败，提示"停机操作中" |
| 锁过期后接管 | 锁已过期 (15 分钟未续期) | 获取锁成功 | B 强制接管，A 后续操作会失败 |
| 强制接管 (--force) | 持有锁 | 强制删除锁并获取 | B 成功接管，A 续期失败后中止 |

#### 3.5.2 锁状态检测

```go
// 检查锁是否过期
func isLockExpired(lock *Lock) bool {
    return time.Since(lock.StartedAt) > 15*time.Minute
}

// 检查当前客户端是否持有锁
func isLockOwner(lock *Lock, clientID string) bool {
    return lock.ClientID == clientID
}
```

### 3.6 幂等性保证

| 操作 | 幂等性保证 | 实现方式 |
|------|-----------|----------|
| deploy (断点续传) | ✓ | 检查已创建资源，跳过已存在步骤 |
| deploy (全新) | ✓ | 检查 state 不存在或 destroyed 才允许 |
| suspend | ✓ | 检查状态，若已 suspended 直接返回成功 |
| resume | ✓ | 检查状态，若已 running 直接返回成功 |
| destroy | ✓ | 检查资源是否存在，已删除的跳过 |
| lock 获取 | ✓ | 使用 If-None-Match 条件写入 |
| state 写入 | ✓ | 使用 If-Match 条件写入（ETag 匹配） |

**deploy 幂等性增强**：
- 不能仅依赖 state 中记录的资源 ID，还要验证云上资源是否真实存在
- 验证 ECS 是否存在、状态是否正常
- 验证 EIP 是否已绑定
- 如果云上资源已被用户手动删除，deploy 应尝试重新创建（而非跳过）

```go
// deploy 跳过逻辑示例
if state.Resources.VPC.ID != "" {
    if _, err := client.DescribeVpc(state.Resources.VPC.ID); err == nil {
        // VPC 存在，跳过创建
        skipCreateVPC = true
    } else {
        // VPC 不存在，需要重新创建
        skipCreateVPC = false
    }
}
```

### 3.7 Edge Case 处理

| 场景 | 检测方式 | 处理策略 |
|------|----------|----------|
| suspend 到一半断电 | 重启检测 `status: suspending` | 查询 ECS 实际状态，Stopped 则更新为 suspended，Running 则回退到 running |
| resume 到一半断电 | 重启检测 `status: resuming` | 查询 ECS 实际状态，Running 则更新为 running，Stopped 则回退到 suspended |
| deploy 到一半断电 | 重启检测 `status: deploying` | 从断点恢复（幂等性保证） |
| destroy 不成功 | 重启检测 `status: destroying` | 保留 state，提示用户重试或放弃 |
| 多进程并发 | 分布式锁 + 状态检查 | 获取锁失败则提示冲突，支持强制接管 |
| 锁过期 | started_at > 15 分钟 | 自动接管，无需用户确认 |
| 续期失败 | RenewLock 返回错误 | 通过 context 取消主操作，保留 state，锁自然过期 |
| OSS 写入失败 | API 返回错误 | 操作中止，已创建资源保留，下次从断点恢复 |
| OSS 不可用 | 请求超时/5xx | 报错退出（v0.1 不支持降级到本地模式） |
| 云上资源被手动删除 | deploy 幂等性检查发现 | 重新创建缺失的资源 |
| 锁文件损坏 | 读取 .lock 解析失败 | 删除损坏的锁文件，重新获取锁 |
| ECS 实际状态与 state 不一致 | API 查询对比 | 以实际状态为准，更新 state.json |
| 快照创建失败 | API 返回错误 | destroy 中断，回退到之前状态 |
| 网络分区 (获取锁后失联) | 锁超时机制 | 15 分钟后其他客户端可强制接管 |

---

## 4. 接口设计

### 4.1 CLI 命令

```
USAGE:
    cloudclaw [COMMAND]

COMMANDS:
    init        配置阿里云凭证并创建 OSS bucket
    deploy      一键部署 OpenClaw
    status      查看部署状态
    destroy     销毁所有云资源（可选保留快照）
    suspend     停机（停止计费，仅收磁盘费）
    resume      恢复运行
    logs        查看容器日志
    ssh         SSH 登录 ECS
    exec        在容器内执行命令
    config      查看或修改配置
    version     显示版本信息
    help        显示帮助信息
```

#### 4.1.1 命令详解

**init**
```bash
cloudclaw init
cloudclaw init --region ap-southeast-1
```

**init 交互流程**：

```
$ cloudclaw init

欢迎使用 CloudClaw！请配置阿里云凭证。

阿里云 AccessKey ID: ***********
阿里云 AccessKey Secret: ***********
区域 [ap-southeast-1]: (回车使用默认值)

正在验证凭证... ✓
正在创建 OSS bucket 'cloudclaw-state-xxx'... ✓
生成 SSH 密钥对... ✓

配置完成！请运行 'cloudclaw deploy' 开始部署。
```

**init 错误处理**：

| 错误场景 | 错误信息 | 恢复建议 |
|----------|----------|----------|
| 凭证无效 | "AccessKey 验证失败，请检查 ID 和 Secret" | 重新输入 |
| 权限不足 | "AccessKey 无 OSS 创建权限，请添加 AliyunOSSFullAccess 策略" | 联系管理员添加权限 |
| Bucket 创建失败 (名称冲突) | "OSS Bucket 名称已被占用，请尝试其他区域" | 检查是否其他账号占用 |
| 网络错误 | "无法连接到阿里云，请检查网络" | 重试或检查网络 |
| 凭证文件写入失败 | "无法写入 credentials 文件，请检查目录权限" | 检查 ~/.cloudclaw 目录权限 |

**deploy**
```bash
cloudclaw deploy                    # 全新部署或从断点恢复
cloudclaw deploy --app              # 仅重新部署应用层
cloudclaw deploy --from-snapshot    # 从快照恢复
cloudclaw deploy --force            # 强制接管（跳过锁检查）
```

**status**
```bash
cloudclaw status
cloudclaw status --json             # JSON 格式输出
```

**destroy**
```bash
cloudclaw destroy                   # 销毁（询问是否保留快照）
cloudclaw destroy --keep-snapshot   # 保留快照
cloudclaw destroy --no-snapshot     # 不保留快照
cloudclaw destroy --force           # 强制销毁（跳过确认）
```

**suspend/resume**
```bash
cloudclaw suspend
cloudclaw resume
```

**logs**
```bash
cloudclaw logs                      # 查看所有容器日志
cloudclaw logs caddy                # 查看 caddy 日志
cloudclaw logs openclaw-gateway     # 查看 gateway 日志
cloudclaw logs -f                   # 跟随输出
```

**ssh**
```bash
cloudclaw ssh           # SSH 登录 ECS
cloudclaw ssh --user root
```

**exec**
```bash
cloudclaw exec openclaw-gateway ls -la
cloudclaw exec caddy cat /etc/caddy/Caddyfile
```

**config**
```bash
cloudclaw config show               # 显示当前配置
cloudclaw config set DOMAIN xxx     # 设置域名
cloudclaw config set GATEWAY_TOKEN xxx
```

### 4.2 与 CloudCode 的差异

| 特性 | CloudCode | CloudClaw |
|------|-----------|-----------|
| 应用容器 | devbox (OpenCode + ttyd) | openclaw-gateway |
| 认证中间件 | Authelia | 无 |
| Web Terminal | ttyd (端口 7681) | 无 |
| 应用端口 | 4096 | 18789 |
| Caddyfile | 复杂 (Authelia + ttyd) | 简单 (直接反向代理) |
| 配置项 | username, passkey | gateway_token |
| logs 命令 | 无 | 有 |
| exec 命令 | 无 | 有 |

---

## 5. 实现规划

### 5.1 文件清单

| 文件 | 来源 | 具体改动 |
|------|------|----------|
| `cmd/cloudclaw/main.go` | 复制修改 | 1. 修改命令名 `cloudcode` → `cloudclaw`<br>2. 修改帮助文本 OpenCode → OpenClaw<br>3. 删除 `ttyd` 子命令<br>4. 新增 `logs`/`exec` 子命令 |
| `internal/alicloud/client.go` | 复制 | 无改动 |
| `internal/alicloud/ecs.go` | 复制 | 无改动 |
| `internal/alicloud/vpc.go` | 复制 | 无改动 |
| `internal/alicloud/eip.go` | 复制 | 无改动 |
| `internal/alicloud/sg.go` | 复制 | 无改动 |
| `internal/alicloud/oss.go` | 复制 | 无改动 |
| `internal/alicloud/sts.go` | 复制 | 无改动 |
| `internal/alicloud/dns.go` | 复制 | 无改动 |
| `internal/alicloud/interfaces.go` | 复制 | 无改动 |
| `internal/alicloud/errors.go` | 复制 | 无改动 |
| `internal/config/credentials.go` | 复制 | 无改动 |
| `internal/config/state.go` | 复制修改 | 1. 删除 `AutheliaConfig` 字段<br>2. 删除 `Passkey` 字段<br>3. 新增 `GatewayToken string` 字段<br>4. 修改 `AppType` 常量值为 "openclaw" |
| `internal/config/backup.go` | 复制 | 无改动 |
| `internal/config/history.go` | 复制 | 无改动 |
| `internal/config/prompt.go` | 复制修改 | 1. 修改提示文本<br>2. 删除 Authelia 相关提示 |
| `internal/deploy/deploy.go` | 复制修改 | 1. 删除 `uploadAutheliaConfig()` 调用<br>2. 删除 `uploadTtydConfig()` 调用<br>3. 修改容器健康检查 URL<br>4. 修改上传的配置文件名 |
| `internal/deploy/destroy.go` | 复制修改 | 1. 修改快照命名前缀<br>2. 修改提示文本 |
| `internal/deploy/suspend.go` | 复制 | 无改动 |
| `internal/deploy/resume.go` | 复制 | 无改动 |
| `internal/deploy/status.go` | 复制修改 | 1. 修改输出格式（OpenClaw 相关信息） |
| `internal/remote/ssh.go` | 复制 | 无改动 |
| `internal/remote/sftp.go` | 复制 | 无改动 |
| `internal/template/render.go` | 复制修改 | 1. 新增 `GatewayToken` 模板变量<br>2. 删除 `Authelia` 相关变量<br>3. 修改模板文件路径 |
| `internal/template/templates/docker-compose.yml.tmpl` | 新建 | 定义 openclaw-gateway + caddy 服务，无 authelia/ttyd 服务 |
| `internal/template/templates/Caddyfile.tmpl` | 新建 | 仅反向代理到 gateway:18789，无 authelia 前置 |
| `internal/template/templates/env.tmpl` | 新建 | OpenClaw 环境变量模板 |
| `tests/unit/*.go` | 复制修改 | 修改测试用例中的服务名称 |
| `tests/e2e/deploy_test.go` | 复制修改 | 修改测试断言（OpenClaw 健康检查） |
| `Makefile` | 复制修改 | 修改二进制名称 `cloudcode` → `cloudclaw` |
| `go.mod` | 复制 | 修改 module 名称 |
| `CLAUDE.md` | 复制 | 无改动 |
| `README.md` | 新建 | CloudClaw 项目说明 |

### 5.2 实现步骤

| 步骤 | 任务 | 依赖 | 预计工作量 |
|------|------|------|-----------|
| 1 | 项目骨架搭建 (go.mod, Makefile, CLAUDE.md) | 无 | 0.5h |
| 2 | 复制 `internal/alicloud/` 模块 | 无 | 0.5h |
| 3 | 复制 `internal/config/` 模块并调整 | 无 | 1h |
| 4 | 复制 `internal/remote/` 模块 | 无 | 0.5h |
| 5 | 复制 `internal/template/` 模块 | 无 | 0.5h |
| 6 | 创建 docker-compose.yml.tmpl | 无 | 0.5h |
| 7 | 创建 Caddyfile.tmpl | 无 | 0.5h |
| 8 | 创建 env.tmpl | 无 | 0.25h |
| 9 | 复制 `internal/deploy/` 模块并调整 | 步骤 2-8 | 2h |
| 10 | 复制 `cmd/` 并修改 | 步骤 9 | 1h |
| 11 | 单元测试 | 步骤 10 | 2h |
| 12 | 端到端测试 | 步骤 11 | 2h |
| 13 | 文档编写 | 步骤 12 | 1h |

**总计**: 约 12 小时

### 5.3 迁移说明

CloudClaw v0.1.0 是首个版本，无历史版本迁移需求。

未来如有版本升级场景，参考 CloudCode v0.3 的迁移方案：
- 检测 state.json 中的版本号
- 根据版本号执行字段迁移/重命名
- 更新 version 字段

---

## 6. 成本影响

新增 OSS bucket 存储：

| 资源 | 月费用 | 备注 |
|------|--------|------|
| OSS 存储（< 1MB） | ~$0.01 | state.json + history |
| OSS 请求费用 | 可忽略 | 低频访问 |
| ECS | ~$20/月 | ecs.e-c1m2.large，按量付费 |
| EIP | ~$2/月 | 按量付费 |
| 停机状态 | ~$3/月 | 仅磁盘费 |

**与 CloudCode 对比**：
- CloudClaw 无需 Authelia 容器，资源占用略低
- 可使用更低配置的 ECS（如 ecs.e-c1m1.large，1C2G）

---

## 7. 测试要点

### 7.1 单元测试

| 模块 | 测试点 |
|------|--------|
| `internal/alicloud/` | 客户端初始化、错误处理、重试机制 |
| `internal/config/` | 凭证读写、状态序列化、权限检查 |
| `internal/template/` | 模板渲染、变量替换、错误处理 |
| `internal/deploy/` | 状态转换逻辑、锁操作、幂等性检查 |

### 7.2 端到端测试

| 场景 | 测试内容 |
|------|----------|
| 全新部署 | `deploy` 创建所有资源，验证健康检查 |
| 断点恢复 | `deploy` 中断后继续，验证资源不重复创建 |
| 停机恢复 | `suspend` → `resume`，验证服务正常 |
| 快照恢复 | `destroy --keep-snapshot` → `deploy --from-snapshot` |
| 并发控制 | 两个终端同时执行 `suspend`，验证锁冲突 |
| 锁过期 | 模拟锁过期后自动接管 |
| 幂等性 | 重复执行 `deploy`，验证不重复创建资源 |
| 强制接管 | `--force` 参数验证 |

### 7.3 手工测试

| 场景 | 测试内容 |
|------|----------|
| 域名访问 | HTTPS 是否正常，证书是否自动续期 |
| Gateway 连接 | WebSocket 是否可用，多频道是否可配置 |
| 日志查看 | `logs` 命令是否正常，`-f` 参数是否有效 |
| SSH 登录 | `ssh` 命令是否正常 |
| 容器执行 | `exec` 命令是否正常 |
| 配置修改 | `config set` 是否生效 |

---

## 变更记录

- v1.3 (2026-03-05): 步骤跳转逻辑修复
  - 修复 suspend 流程失败分支跳转到步骤 10（释放锁）
  - 修复 resume 流程失败分支跳转到步骤 11（释放锁）
  - 修复 destroy 流程失败分支跳转到步骤 10（释放锁）
  - 统一图例说明格式


- v1.2 (2026-03-05): 第二轮评审修复
  - 补充 deploy 流程带锁版本（锁冲突处理、续期机制、强制接管）
  - 补充 suspend/resume/destroy 流程失败分支处理
  - 补充断点检查位置说明（获取锁之前）
  - 补充 CHECKPOINT 断点标记

- v1.1 (2026-03-05): 第一轮评审修复
  - 补充状态存储方案对比（2.3 节）
  - 补充本地 vs OSS 职责划分表
  - 补充 previous_status 字段详细说明
  - 补充完整的 deploy/suspend/resume/destroy 操作流程
  - 补充并发控制场景设计（3.5 节）
  - 补充幂等性保证表格（3.6 节）
  - 补充 Edge Case 处理表格（3.7 节）
  - 补充 init 命令详细交互流程和错误处理
  - 细化文件清单（具体到改动点）

- v1.0 (2026-03-05): 初始版本，基于 CloudCode v0.3 架构设计
