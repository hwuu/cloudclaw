# CloudClaw 安装指南

本文档介绍如何安装和配置 CloudClaw。

## 系统要求

- **操作系统**: Linux / macOS / WSL2
- **Go 版本**: Go 1.26+
- **Git**: 用于克隆代码仓库

## 安装步骤

### 1. 克隆仓库

```bash
git clone https://github.com/hwuu/cloudclaw.git
cd cloudclaw
```

### 2. 构建二进制文件

```bash
make build
```

构建成功后，二进制文件位于 `bin/cloudclaw`。

### 3. 配置环境变量

CloudClaw 需要阿里云访问密钥才能工作。设置以下环境变量：

```bash
export ALICLOUD_ACCESS_KEY_ID="your_access_key_id"
export ALICLOUD_ACCESS_KEY_SECRET="your_access_key_secret"
```

或者，创建配置文件 `~/.alicloud/config`：

```ini
[default]
access_key_id = your_access_key_id
access_key_secret = your_access_key_secret
region = ap-southeast-1
```

### 4. 验证安装

```bash
./bin/cloudclaw version
```

应该输出版本信息：

```
cloudclaw main
  commit: <commit-hash>
  built:  <build-time>
  go:     go1.26.0
```

## 快速开始

安装完成后，运行以下命令一键部署 OpenClaw：

```bash
./bin/cloudclaw deploy
```

部署完成后，可以使用以下命令管理：

```bash
# 查看部署状态
./bin/cloudclaw status

# SSH 登录 ECS
./bin/cloudclaw ssh

# 停机（停止计费，仅收磁盘费）
./bin/cloudclaw suspend

# 恢复运行
./bin/cloudclaw resume

# 销毁所有资源
./bin/cloudclaw destroy
```

## 命令概览

| 命令 | 说明 |
|------|------|
| `cloudclaw deploy` | 一键部署 OpenClaw |
| `cloudclaw deploy --app` | 仅重部署应用层（不创建云资源） |
| `cloudclaw status` | 查看部署状态 |
| `cloudclaw destroy` | 销毁所有云资源 |
| `cloudclaw suspend` | 停机（停止计费） |
| `cloudclaw resume` | 恢复运行 |
| `cloudclaw ssh` | SSH 登录 ECS |
| `cloudclaw exec <cmd>` | 在容器中执行命令 |
| `cloudclaw plugins` | 插件管理 |
| `cloudclaw channels` | 渠道管理 |
| `cloudclaw version` | 显示版本信息 |

## 插件管理

```bash
# 列出可用插件
./bin/cloudclaw plugins list

# 安装插件
./bin/cloudclaw plugins install feishu
./bin/cloudclaw plugins install telegram

# 启用/禁用插件
./bin/cloudclaw plugins enable feishu
./bin/cloudclaw plugins disable feishu

# 卸载插件
./bin/cloudclaw plugins uninstall feishu
```

## 渠道配置

```bash
# 列出已配置渠道
./bin/cloudclaw channels list

# 添加飞书渠道
./bin/cloudclaw channels add my-feishu \
  --type feishu \
  --webhook-url "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"

# 添加 Telegram 渠道
./bin/cloudclaw channels add my-telegram \
  --type telegram \
  --bot-token "xxx:yyy" \
  --chat-id "123456"

# 启用/禁用渠道
./bin/cloudclaw channels enable my-feishu
./bin/cloudclaw channels disable my-feishu

# 删除渠道
./bin/cloudclaw channels remove my-feishu
```

## 故障排查

### 配置加载失败

确保已设置阿里云访问密钥：

```bash
echo $ALICLOUD_ACCESS_KEY_ID
echo $ALICLOUD_ACCESS_KEY_SECRET
```

### SSH 连接失败

检查 ECS 实例是否运行正常，并且安全组已开放 22 端口。

### 命令不识别

确保使用 `./bin/cloudclaw` 或在 PATH 中添加二进制文件：

```bash
export PATH="$PATH:$PWD/bin"
```

## 从源码安装到 PATH

```bash
# 构建并复制到 PATH
make build
sudo cp bin/cloudclaw /usr/local/bin/

# 验证
cloudclaw version
```

## 更新

```bash
git pull
make build
./bin/cloudclaw version
```

## 卸载

```bash
# 如果复制到了 /usr/local/bin
sudo rm /usr/local/bin/cloudclaw

# 删除状态目录
rm -rf ~/.cloudclaw
```

## 获取帮助

```bash
# 查看所有命令
./bin/cloudclaw --help

# 查看具体命令帮助
./bin/cloudclaw deploy --help
./bin/cloudclaw plugins --help
```
