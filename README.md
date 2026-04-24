# Bangumi-PikPak

Bangumi-PikPak 是基于 [Mikan Project](https://mikanani.me) 和 [PikPak](https://mypikpak.com/) 的自动追番整理下载工具。订阅 Mikan RSS 后，程序会定时检测新番剧条目，下载 torrent，并提交到 PikPak 离线下载。

当前主实现已迁移为 Go。原 Python 单文件实现 `main.py` 暂时保留为 legacy reference，方便对照和回退。

## 功能特性

- 兼容现有 `config.json` 配置格式
- 定时解析 Mikan RSS
- 自动访问 Mikan 条目页提取番剧标题
- 本地缓存 torrent 到 `torrent/<番剧名>/`
- 自动登录 PikPak
- 自动创建 PikPak 番剧文件夹
- 提交 PikPak 离线下载任务
- 尽量避免重复提交
- 支持 HTTP/HTTPS/SOCKS 代理环境变量
- 控制台日志与 `rss-pikpak.log` 轮转日志
- 支持单次运行、持续运行、Docker 和 systemd 部署

## 环境要求

- Go 1.22 或更高版本
- 有效 PikPak 账号
- Mikan Project RSS 订阅链接

Go 版使用 [`github.com/kanghengliu/pikpak-go`](https://pkg.go.dev/github.com/kanghengliu/pikpak-go) 访问 PikPak。

## 配置

复制示例配置：

```bash
cp example.config.json config.json
```

编辑 `config.json`：

```json
{
    "username": "your_email@example.com",
    "password": "your_password",
    "path": "your_pikpak_folder_id",
    "rss": "https://mikanani.me/RSS/MyBangumi?token=your_token_here",
    "http_proxy": "http://127.0.0.1:7890",
    "https_proxy": "http://127.0.0.1:7890",
    "socks_proxy": "socks5://127.0.0.1:7890",
    "enable_proxy": false
}
```

配置项：

| 配置项 | 说明 |
|---|---|
| `username` | PikPak 账号邮箱 |
| `password` | PikPak 账号密码 |
| `path` | PikPak 目标父文件夹 ID |
| `rss` | Mikan RSS 订阅链接 |
| `http_proxy` | HTTP 代理地址 |
| `https_proxy` | HTTPS 代理地址 |
| `socks_proxy` | SOCKS 代理地址 |
| `enable_proxy` | 是否启用代理 |

> `pikpak.json` 在 Go 版中作为运行状态文件使用。由于 Python 版 `pikpakapi` 的 token 序列化格式和 `pikpak-go` 的公开 API 不同，Go 版不保证复用旧 Python token；需要时会用账号密码重新登录。

## Go 版本运行

构建：

```bash
go build ./cmd/bangumi-pikpak
```

运行：

```bash
./bangumi-pikpak -config config.json
```

Windows PowerShell：

```powershell
.\bangumi-pikpak.exe -config config.json
```

单次运行，适合测试、cron 或容器任务：

```bash
go run ./cmd/bangumi-pikpak -config config.json -once
```

## 常用参数

- `-config config.json`：配置文件路径
- `-interval 600`：RSS 检查间隔，单位秒
- `-once`：只运行一次后退出
- `-log rss-pikpak.log`：日志文件路径
- `-state pikpak.json`：运行状态文件路径

## Docker

构建镜像：

```bash
docker build -t bangumi-pikpak-go .
```

运行容器：

```bash
docker run -d \
  --name bangumi-pikpak \
  -v /path/to/data:/app/data \
  bangumi-pikpak-go
```

确保 `/path/to/data/config.json` 已存在。

## systemd

示例服务文件位于：

```text
docs/examples/bangumi-pikpak.service
```

典型部署路径：

```bash
sudo mkdir -p /opt/bangumi-pikpak
sudo cp bangumi-pikpak config.json /opt/bangumi-pikpak/
sudo cp docs/examples/bangumi-pikpak.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now bangumi-pikpak
```

## 工作流程

1. 读取 `config.json`。
2. 应用代理配置。
3. 拉取并解析 Mikan RSS。
4. 访问条目页面读取 `p.bangumi-title`。
5. 检查本地 torrent 缓存。
6. 有新 torrent 时登录 PikPak。
7. 创建或复用番剧文件夹。
8. 下载 torrent 到本地。
9. 提交 PikPak 离线下载任务。
10. 持续运行时等待下一个检查周期。

## 开发验证

```bash
go test ./...
go build ./cmd/bangumi-pikpak
go run ./cmd/bangumi-pikpak -config example.config.json -once
```

`example.config.json` 只包含示例值，不能用于真实下载。

## 许可证

MIT，详见 `LICENSE`。
