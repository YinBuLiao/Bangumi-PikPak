# Bangumi-PikPak Linux 部署指南

**运行环境**：Ubuntu 24.04 64位

## 第一步：配置 Mihomo 代理服务

本部署方案使用 Mihomo 作为代理客户端，确保网络连接稳定。

创建 Mihomo 系统服务配置文件：

```bash
sudo nano /etc/systemd/system/mihomo.service
```

在打开的编辑器中写入以下内容：

```ini
[Unit]
Description=Mihomo Proxy Service
After=network.target

[Service]
# User/Group: 建议为安全起见，使用一个专用的、无特权的用户来运行此服务。
# 例如，创建一个名为 'mihomo_user' 的用户，并设置 '/mihomo' 目录的所有权归其所有。
# User=mihomo_user
# Group=mihomo_user
# 如果以 root 用户运行（为了简化操作，但安全性较低）：
User=root
Group=root

# WorkingDirectory: Mihomo 可执行文件及其配置文件所在的目录
WorkingDirectory=/mihomo/
# ExecStart: 执行 Mihomo 的命令
# 确保可执行文件的名称与您的文件相符：mihomo-linux-amd64-v1.18.1
ExecStart=/mihomo/mihomo-linux-amd64-v1.18.1 -d .

# Restart Policy: 如果 Mihomo 崩溃，则重新启动它
Restart=always
RestartSec=5s

# Logging: Direct output to journalctl
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 1.1 重新加载 systemd 配置

```bash
sudo systemctl daemon-reload
```

### 1.2 启用 Mihomo 服务（开机自启动）

```bash
sudo systemctl enable mihomo.service
```

### 1.3 启动 Mihomo 服务

```bash
sudo systemctl start mihomo.service
```

### 1.4 验证 Mihomo 服务状态

查看服务状态和实时日志：

```bash
sudo systemctl status mihomo.service
sudo journalctl -f -u mihomo.service
```

## 第二步：配置 Bangumi-PikPak 服务

**前置条件**：请确保已在项目目录下创建 Python 虚拟环境（建议命名为 `venv`）并安装所需依赖。

创建 Bangumi-PikPak 系统服务配置文件：

```bash
sudo nano /etc/systemd/system/bangumi-pikpak.service
```

在打开的编辑器中写入以下内容：

```ini
[Unit]
Description=Bangumi PikPak Service
# 确保网络已经就绪，并且 mihomo.service 在此服务启动之前已经启动。
After=network.target mihomo.service
Requires=mihomo.service

[Service]
# IMPORTANT: 建议为安全起见，使用一个专用的、无特权的用户来运行此服务。
User=root
Group=root

# WorkingDirectory: 您的 Python 脚本的工作目录（main.py 和 venv 所在的位置）
WorkingDirectory=/root/Bangumi-PikPak/

# Proxy Environment Variables: 代理环境变量
Environment="HTTP_PROXY=http://127.0.0.1:7890"
Environment="HTTPS_PROXY=http://127.0.0.1:7890"
Environment="ALL_PROXY=socks5://127.0.0.1:7890"

# Execute Python Script from Virtual Environment: 从虚拟环境执行 Python 脚本
# 使用虚拟环境内部 Python 可执行文件的完整路径。
ExecStart=/root/Bangumi-PikPak/venv/bin/python main.py

Restart=always
RestartSec=5s

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target

```

### 2.1 重新加载 systemd 配置

```bash
sudo systemctl daemon-reload
```

### 2.2 启用 Bangumi-PikPak 服务（开机自启动）

```bash
sudo systemctl enable bangumi-pikpak.service
```

### 2.3 启动 Bangumi-PikPak 服务

```bash
sudo systemctl start bangumi-pikpak.service
```

### 2.4 验证 Bangumi-PikPak 服务状态

查看服务状态和实时日志：

```bash
sudo systemctl status bangumi-pikpak.service
sudo journalctl -f -u bangumi-pikpak.service
```

---

## 部署完成

至此，Bangumi-PikPak 已成功部署为 Linux 系统服务。服务将在系统启动时自动运行，并具备自动重启机制。

**注意事项**：
- 确保 Mihomo 代理服务正常运行，Bangumi-PikPak 依赖它进行网络访问
- 如需修改配置，请编辑相应的配置文件后重启服务
- 可通过 `journalctl` 命令查看详细的运行日志进行故障排查

