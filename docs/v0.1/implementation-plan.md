# CloudClaw v0.1.0 实现计划

## 项目概述

CloudClaw 是一个类似 CloudCode 的 CLI 工具，用于一键部署 OpenClaw 到阿里云 ECS。

**核心目标**：复用 CloudCode 80% 代码，快速实现 OpenClaw 部署。

---

## 实现阶段

### 阶段 1: 项目骨架搭建 (P0) ✅ 已完成

**目标**：建立项目基础结构，可编译运行

**任务**：
1. 创建 `go.mod` (修改 module 名称) ✅
2. 创建 `Makefile` (修改二进制名称) ✅
3. 创建 `CLAUDE.md` (复制 CloudCode) ✅
4. 创建 `README.md` (新建项目说明) ✅
5. 创建 `cmd/cloudclaw/main.go` (CLI 入口，先实现 version 命令) ✅

**交付物**：
- [x] `go.mod`
- [x] `Makefile`
- [x] `CLAUDE.md`
- [x] `README.md`
- [x] `cmd/cloudclaw/main.go`

**验收标准**：
- `make build` 可成功编译 ✅
- `./bin/cloudclaw version` 输出版本信息 ✅

**评审结果**：
- GLM-5: 8/10 - 无 Critical
- Kimi-K2.5: 8.5/10 - 无 Critical
- MiniMax-M2.5: 9/10 - 无 Critical

**Git 提交**：
- `fd0be92` - chore: 初始化项目骨架
- `424b113` - chore: 添加.gitignore 忽略二进制产物
- `6574f4c` - chore: 修复评审问题
- `d2a08bf` - refactor: main.go 注释改回中文，与 CloudCode 保持一致
- `428f492` - refactor: 修复评审 Warning 问题
- `40694eb` - refactor: 修复 args 未使用 warning
- `b4d2b77` - chore: go 版本改回 1.26.0（当前稳定版）

---

### 阶段 2: 阿里云 SDK 封装 (P0)

**目标**：复用 CloudCode 的阿里云 SDK 封装

**任务**：
1. 复制 `internal/alicloud/` 目录
2. 修改包名导入路径
3. 验证代码编译通过

**交付物**：
- [ ] `internal/alicloud/client.go`
- [ ] `internal/alicloud/ecs.go`
- [ ] `internal/alicloud/vpc.go`
- [ ] `internal/alicloud/eip.go`
- [ ] `internal/alicloud/sg.go` (安全组)
- [ ] `internal/alicloud/oss.go`
- [ ] `internal/alicloud/sts.go`
- [ ] `internal/alicloud/dns.go`
- [ ] `internal/alicloud/interfaces.go`
- [ ] `internal/alicloud/errors.go`

**验收标准**：
- 所有文件复制完成
- 包名和导入路径修改正确
- `make build` 编译通过

---

### 阶段 3: 配置管理模块 (P0)

**目标**：复用 CloudCode 的配置管理，调整为 OpenClaw 配置

**任务**：
1. 复制 `internal/config/` 目录
2. 修改 `state.go` (去掉 Authelia 相关字段，增加 GatewayToken)
3. 修改 `prompt.go` (修改提示文本)
4. 删除/修改 Authelia 相关文件

**交付物**：
- [ ] `internal/config/credentials.go`
- [ ] `internal/config/state.go` (修改)
- [ ] `internal/config/backup.go`
- [ ] `internal/config/history.go`
- [ ] `internal/config/prompt.go` (修改)

**验收标准**：
- 配置结构体正确 (无 Authelia 字段)
- 编译通过

---

### 阶段 4: 模板渲染模块 (P0)

**目标**：创建 OpenClaw 专用的模板文件

**任务**：
1. 复制 `internal/template/render.go`
2. 创建 `docker-compose.yml.tmpl` (OpenClaw 配置)
3. 创建 `Caddyfile.tmpl` (简化版，无 Authelia)
4. 创建 `env.tmpl` (环境变量模板)

**交付物**：
- [ ] `internal/template/render.go`
- [ ] `internal/template/templates/docker-compose.yml.tmpl`
- [ ] `internal/template/templates/Caddyfile.tmpl`
- [ ] `internal/template/templates/env.tmpl`

**验收标准**：
- 模板文件语法正确
- 模板变量与配置结构匹配

---

### 阶段 5: 远程连接模块 (P0)

**目标**：复用 CloudCode 的 SSH/SFTP 功能

**任务**：
1. 复制 `internal/remote/` 目录
2. 修改包名导入路径

**交付物**：
- [ ] `internal/remote/ssh.go`
- [ ] `internal/remote/sftp.go`

**验收标准**：
- 编译通过

---

### 阶段 6: 部署编排模块 (P0)

**目标**：复用 CloudCode 的部署流程，修改应用层部署

**任务**：
1. 复制 `internal/deploy/` 目录
2. 修改 `deploy.go` (应用部署部分，改为 OpenClaw)
3. 修改 `destroy.go` (修改快照命名等)
4. 修改 `status.go` (修改输出格式)
5. 保持 `suspend.go` 和 `resume.go` 不变

**交付物**：
- [ ] `internal/deploy/deploy.go` (修改)
- [ ] `internal/deploy/destroy.go` (修改)
- [ ] `internal/deploy/suspend.go`
- [ ] `internal/deploy/resume.go`
- [ ] `internal/deploy/status.go` (修改)
- [ ] `internal/deploy/dns.go`

**验收标准**：
- 部署流程正确
- 应用层部署针对 OpenClaw

---

### 阶段 7: CLI 命令实现 (P0)

**目标**：实现完整的 CLI 命令

**任务**：
1. 完善 `cmd/cloudclaw/main.go`
2. 实现 `init` 命令
3. 实现 `deploy` 命令
4. 实现 `status` 命令
5. 实现 `destroy` 命令
6. 实现 `suspend` 命令
7. 实现 `resume` 命令
8. 实现 `logs` 命令
9. 实现 `ssh` 命令
10. 实现 `exec` 命令
11. 实现 `config` 命令
12. 实现 `version` 命令

**交付物**：
- [ ] `cmd/cloudclaw/main.go` (完整)

**验收标准**：
- 所有命令可正常执行
- 帮助文本正确

---

### 阶段 8: 插件管理命令 (P1)

**目标**：实现 OpenClaw 插件管理功能

**任务**：
1. 实现 `plugins list` 命令
2. 实现 `plugins install` 命令
3. 实现 `plugins uninstall` 命令
4. 实现 `plugins enable` 命令

**交付物**：
- [ ] `internal/deploy/plugins.go`
- [ ] `cmd/cloudclaw/main.go` (plugins 子命令)

**验收标准**：
- 可通过 SSH 安装/卸载插件
- 支持飞书、Telegram 等插件

---

### 阶段 9: 渠道配置命令 (P1)

**目标**：实现聊天渠道配置功能

**任务**：
1. 实现 `channels add` 命令
2. 实现 `channels list` 命令
3. 实现 `channels remove` 命令
4. 交互式配置流程

**交付物**：
- [ ] `internal/deploy/channels.go`
- [ ] `cmd/cloudclaw/main.go` (channels 子命令)

**验收标准**：
- 支持飞书、Telegram 等渠道配置
- 配置后自动重启 Gateway

---

### 阶段 10: 单元测试 (P1)

**目标**：编写单元测试保证质量

**任务**：
1. 创建 `tests/unit/` 目录
2. 编写配置模块测试
3. 编写部署模块测试
4. 编写模板渲染测试

**交付物**：
- [ ] `tests/unit/main_test.go`
- [ ] `tests/unit/config_test.go`
- [ ] `tests/unit/template_test.go`
- [ ] `tests/unit/deploy_test.go`

**验收标准**：
- `make test` 运行通过
- 关键逻辑有测试覆盖

---

### 阶段 11: 端到端测试 (P2)

**目标**：验证完整部署流程

**任务**：
1. 创建 `tests/e2e/` 目录
2. 编写部署测试
3. 编写 suspend/resume 测试
4. 编写 destroy 测试

**交付物**：
- [ ] `tests/e2e/deploy_test.go`
- [ ] `tests/e2e/suspend_resume_test.go`
- [ ] `tests/e2e/destroy_test.go`

**验收标准**：
- 端到端测试通过

---

### 阶段 12: 文档编写 (P2)

**目标**：完善项目文档

**任务**：
1. 完善 `README.md`
2. 编写 `INSTALL.md` (安装指南)
3. 编写 `USAGE.md` (使用指南)
4. 更新设计文档

**交付物**：
- [ ] `README.md`
- [ ] `docs/INSTALL.md`
- [ ] `docs/USAGE.md`

**验收标准**：
- 文档完整清晰

---

## 实现顺序依赖

```
阶段 1 (骨架) ✅
    ↓
阶段 2 (alicloud) → 阶段 3 (config) → 阶段 4 (template) → 阶段 5 (remote)
                                      ↓
                                    阶段 6 (deploy)
                                      ↓
                                    阶段 7 (CLI)
                                      ↓
                    阶段 8 (plugins) → 阶段 9 (channels)
                                      ↓
                                    阶段 10 (测试)
                                      ↓
                                    阶段 11 (E2E)
                                      ↓
                                    阶段 12 (文档)
```

---

## 代码复用策略

| 模块 | 复用方式 | 修改内容 |
|------|----------|----------|
| `internal/alicloud/` | 100% 复制 | 无 |
| `internal/config/` | 复制后修改 | 去掉 Authelia 相关字段 |
| `internal/deploy/` | 复制后修改 | 应用部署部分改为 OpenClaw |
| `internal/template/` | 复制框架 | 模板内容全新编写 |
| `internal/remote/` | 100% 复制 | 无 |
| `cmd/` | 复制后修改 | 命令名称、帮助文本、新增 plugins/channels |

---

## 预计工作量

| 阶段 | 预计工时 | 优先级 |
|------|----------|--------|
| 阶段 1 | 0.5h | P0 |
| 阶段 2 | 0.5h | P0 |
| 阶段 3 | 1h | P0 |
| 阶段 4 | 1h | P0 |
| 阶段 5 | 0.5h | P0 |
| 阶段 6 | 2h | P0 |
| 阶段 7 | 2h | P0 |
| 阶段 8 | 1.5h | P1 |
| 阶段 9 | 1.5h | P1 |
| 阶段 10 | 2h | P1 |
| 阶段 11 | 2h | P2 |
| 阶段 12 | 1h | P2 |

**总计**: 约 16 小时

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| OpenClaw 配置变化 | 模板可能不适用 | 使用环境变量，便于调整 |
| 阿里云 API 变更 | SDK 可能需要更新 | 跟进 CloudCode 更新 |
| 插件 API 不稳定 | 插件管理可能失效 | 封装 SSH 调用，隔离变化 |

---

## 下一步行动

1. **确认实现计划** - 用户审核本计划 ✅
2. **开始阶段 1** - 项目骨架搭建 ✅
3. **Quorum Review** - 每阶段完成后评审 ✅
4. **用户审核** - 关键阶段后人工审核 ✅
5. **开始阶段 2** - 阿里云 SDK 封装 (下一步)
