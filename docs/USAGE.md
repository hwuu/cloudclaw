# CloudClaw 使用指南

本文档介绍 CloudClaw 的详细用法。

## 目录

1. [快速开始](#快速开始)
2. [部署应用](#部署应用)
3. [管理资源](#管理资源)
4. [插件管理](#插件管理)
5. [渠道配置](#渠道配置)
6. [常见问题](#常见问题)

---

## 快速开始

### 1. 配置阿里云凭证

```bash
export ALICLOUD_ACCESS_KEY_ID="your_access_key_id"
export ALICLOUD_ACCESS_KEY_SECRET="your_access_key_secret"
```

### 2. 一键部署

```bash
./bin/cloudclaw deploy
```

部署过程：
1. 创建 VPC 和交换机
2. 创建 ECS 实例
3. 分配公网 IP
4. 配置域名解析
5. 部署 Docker 和 OpenClaw

部署完成后，输出包含访问地址。

---

## 部署应用

### 首次部署

```bash
./bin/cloudclaw deploy
```

### 重部署应用层

如果只更新应用配置（不改变云资源）：

```bash
./bin/cloudclaw deploy --app
```

### 部署选项

| 选项 | 说明 |
|------|------|
| `--app` | 仅部署应用层，不创建云资源 |
| `--region <region>` | 指定阿里云区域 |

---

## 管理资源

### 查看状态

```bash
./bin/cloudclaw status
```

输出包含：
- ECS 实例状态
- 公网 IP
- 容器运行状态
- 域名配置

### 停机/恢复

**停机**（停止计费，仅收磁盘费）：

```bash
./bin/cloudclaw suspend
```

**恢复运行**：

```bash
./bin/cloudclaw resume
```

### 销毁资源

**预览删除**（不实际执行）：

```bash
./bin/cloudclaw destroy --dry-run
```

**跳过确认删除**：

```bash
./bin/cloudclaw destroy --force
```

**完整删除**：

```bash
./bin/cloudclaw destroy
```

---

## SSH 访问

### 登录 ECS

```bash
./bin/cloudclaw ssh
```

### 在容器中执行命令

```bash
# 默认在 devbox 容器中
./bin/cloudclaw exec "ls -la /app"

# 指定容器
./bin/cloudclaw exec --container devbox "docker ps"
```

---

## 插件管理

CloudClaw 支持通过插件扩展 OpenClaw 功能。

### 支持的插件

| 插件 | 说明 |
|------|------|
| feishu | 飞书消息通知 |
| telegram | Telegram 消息通知 |
| discord | Discord 消息通知 |
| wechat | 企业微信消息通知 |

### 列出插件

```bash
./bin/cloudclaw plugins list
```

输出示例：

```
名称         描述                           版本       状态
----------------------------------------------------------------------
feishu       飞书消息通知插件                  latest     已启用
telegram     Telegram 消息通知插件             latest     未安装
discord      Discord 消息通知插件              latest     未安装
wechat       企业微信消息通知插件              latest     未安装
```

### 安装插件

```bash
./bin/cloudclaw plugins install feishu
```

### 启用/禁用插件

```bash
./bin/cloudclaw plugins enable feishu
./bin/cloudclaw plugins disable feishu
```

### 卸载插件

```bash
./bin/cloudclaw plugins uninstall feishu
```

---

## 渠道配置

配置消息通知渠道，OpenClaw 将通过这些渠道发送通知。

### 支持的渠道

| 渠道 | 必需参数 |
|------|----------|
| feishu | --webhook-url |
| telegram | --bot-token, --chat-id |
| discord | --webhook-url |
| wechat | --webhook-url |

### 列出渠道

```bash
./bin/cloudclaw channels list
```

输出示例：

```
名称              类型         配置                   状态
----------------------------------------------------------------------
my-feishu      feishu     {"webhook_url":"htt... 已启用
```

### 添加飞书渠道

1. 在飞书创建机器人，获取 Webhook URL
2. 添加渠道：

```bash
./bin/cloudclaw channels add my-feishu \
  --type feishu \
  --webhook-url "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
```

### 添加 Telegram 渠道

1. 联系 @BotFather 创建 Bot，获取 Token
2. 联系 @userinfobot 获取 Chat ID
3. 添加渠道：

```bash
./bin/cloudclaw channels add my-telegram \
  --type telegram \
  --bot-token "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11" \
  --chat-id "123456789"
```

### 添加 Discord 渠道

1. 在 Discord 服务器创建 Webhook
2. 添加渠道：

```bash
./bin/cloudclaw channels add my-discord \
  --type discord \
  --webhook-url "https://discord.com/api/webhooks/xxx/yyy"
```

### 添加企业微信渠道

1. 在企业微信创建机器人，获取 Webhook URL
2. 添加渠道：

```bash
./bin/cloudclaw channels add my-wechat \
  --type wechat \
  --webhook-url "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
```

### 启用/禁用渠道

```bash
./bin/cloudclaw channels enable my-feishu
./bin/cloudclaw channels disable my-feishu
```

### 删除渠道

```bash
./bin/cloudclaw channels remove my-feishu
```

---

## 常见问题

### 1. 部署失败怎么办？

检查以下几点：
- 阿里云凭证是否正确
- 账户余额是否充足
- 区域是否有可用资源

查看详细错误信息：

```bash
./bin/cloudclaw deploy 2>&1 | tee deploy.log
```

### 2. 如何修改配置？

编辑 OSS 上的状态文件，然后重新部署：

```bash
./bin/cloudclaw deploy --app
```

### 3. 停机后费用是多少？

停机状态下：
- ECS 实例：停止计费
- 公网 IP：停止计费
- 磁盘：继续计费（约 $3/月）

### 4. 如何备份数据？

使用快照功能：

```bash
# 通过阿里云控制台创建快照
# 或使用 CLI
aliyun ecs CreateSnapshot --InstanceId i-xxx
```

### 5. 忘记 SSH 私钥位置？

SSH 私钥存储在：

```bash
ls ~/.cloudclaw/ssh_key
```

### 6. 如何重置部署？

```bash
# 销毁所有资源
./bin/cloudclaw destroy --force

# 清理状态
rm -rf ~/.cloudclaw

# 重新部署
./bin/cloudclaw deploy
```

### 7. 域名解析不生效？

等待 DNS 传播（通常 5-10 分钟），或检查：
- 域名是否正确配置
- 域名服务器是否生效

```bash
dig your-domain.com
```

### 8. 如何查看日志？

```bash
# SSH 登录后查看
./bin/cloudclaw ssh
docker logs devbox
```

---

## 技术支持

遇到问题？

1. 查看 [README.md](../README.md)
2. 查看 [INSTALL.md](INSTALL.md)
3. 提交 GitHub Issue
