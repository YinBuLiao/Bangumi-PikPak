# Bangumi-PikPak

> ⚠️ **重要声明**：本项目是基于 [YinBuLiao](https://github.com/YinBuLiao) 的 [Bangumi-PikPak](https://github.com/YinBuLiao/Bangumi-PikPak) 项目的改进版本，主要添加了代理支持等功能。

<p align="center">
    <img title="mikan project" src="https://mikanani.me/images/mikan-pic.png" alt="" width="10%">
    <img title="pikpak" src="https://raw.githubusercontent.com/YinBuLiao/Bangumi-PikPak/main/img/pikpak.png">
</p>

***

## 🔗 与原项目的关系

- **原项目**：[YinBuLiao/Bangumi-PikPak](https://github.com/YinBuLiao/Bangumi-PikPak)
- **改进内容**：添加代理支持、改进配置管理、完善文档等
- **许可证**：遵循原项目的 MIT 许可证

## ✨ 新增功能

- 🌐 **代理支持**：支持 HTTP/HTTPS/SOCKS 代理
- ⚙️ **改进配置**：更好的配置文件管理
- 📊 **完善日志**：详细的日志记录系统
- 📚 **项目文档**：完整的安装和使用说明

---

本项目是基于 [Mikan Project](https://mikanani.me)、[PikPak](https://mypikpak.com/) 的全自动追番整理下载工具。只需要在 [Mikan Project](https://mikanani.me) 上订阅番剧，就可以全自动追番。

## ✨ 功能特性

- 🚀 **简易配置**：单次配置就能持续使用
- 🔄 **自动更新**：无需介入的 RSS 解析器，自动解析番组信息
- 📁 **智能整理**：根据番剧更新时间自动分类整理
- 🤖 **完全自动化**：无需维护，完全无感使用
- 🌐 **代理支持**：支持 HTTP/HTTPS/SOCKS 代理
- 📊 **日志记录**：完整的操作日志，便于问题排查

## 🚀 快速开始

### 环境要求

- Python 3.10 或更高版本
- 有效的 PikPak 账号
- Mikan Project 的 RSS 订阅链接

### 安装步骤

1. **克隆项目**
```bash
git clone https://github.com/hrWong/Bangumi-PikPak.git
cd Bangumi-PikPak
```

2. **安装依赖**
```bash
pip install -r requirements.txt
```

3. **配置设置**
```bash
# 复制示例配置文件
cp example.config.json config.json

# 编辑配置文件
# 填入你的 PikPak 账号信息和 RSS 链接
```

4. **运行程序**
```bash
python main.py
```

## ⚙️ 配置说明

### 配置文件格式

编辑 `config.json` 文件：

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

### 配置项说明

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `username` | PikPak 账号邮箱 | `user@example.com` |
| `password` | PikPak 账号密码 | `your_password` |
| `path` | PikPak 目标文件夹ID | `VOXXWeEex835fv5C2hV5LBe1o2` |
| `rss` | Mikan RSS 订阅链接 | `https://mikanani.me/RSS/MyBangumi?token=xxx` |
| `http_proxy` | HTTP 代理地址 | `http://127.0.0.1:7890` |
| `https_proxy` | HTTPS 代理地址 | `http://127.0.0.1:7890` |
| `socks_proxy` | SOCKS 代理地址 | `socks5://127.0.0.1:7890` |
| `enable_proxy` | 是否启用代理 | `true` 或 `false` |

### 获取配置信息

#### 文件夹ID
1. 登录 [PikPak](https://mypikpak.com/)
2. 创建或选择目标文件夹
3. 从URL中复制文件夹ID：`https://mypikpak.com/drive/folder/文件夹ID`

#### RSS链接
1. 在 [Mikan Project](https://mikanani.me) 订阅番剧
2. 在首页右下角复制 RSS 订阅链接
3. 格式：`https://mikanani.me/RSS/MyBangumi?token=xxx%3d%3d`

## 📦 打包部署

### 生成可执行文件

```bash
# 安装 PyInstaller
pip install pyinstaller

# 打包程序
pyinstaller --onefile --noconsole main.py
```

### 开机自启动

#### Windows
- 将生成的可执行文件放入启动文件夹
- 或创建任务计划程序

#### macOS
- 系统偏好设置 → 用户与群组 → 登录项 → 添加可执行文件

#### Linux
- 将可执行文件路径添加到 `~/.bashrc`
- 或创建 systemd 服务

## 🔧 工作原理

### 更新检测流程
1. **RSS 解析**：定期检查 Mikan RSS 源
2. **番剧识别**：访问番剧页面提取标题信息
3. **文件夹管理**：自动创建番剧分类文件夹
4. **种子处理**：下载种子并上传到 PikPak
5. **重复检测**：智能避免重复内容

### 文件组织结构
```
PikPak 根目录/
├── 进击的巨人 最终季/
│   ├── 第1集.torrent
│   └── 第2集.torrent
├── 阿松 第四季/
│   ├── 第1集.torrent
│   └── 第2集.torrent
└── ...
```

## 📝 日志说明

程序运行时会生成详细的日志文件 `rss-pikpak.log`，包含：
- 配置加载状态
- 代理设置信息
- RSS 更新检测
- 文件操作记录
- 错误和警告信息

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发环境设置
```bash
# 克隆项目
git clone https://github.com/your-username/Bangumi-PikPak.git
cd Bangumi-PikPak

# 创建虚拟环境
python -m venv venv
source venv/bin/activate  # Linux/macOS
# 或
venv\Scripts\activate  # Windows

# 安装开发依赖
pip install -r requirements.txt
```

### 代码规范
- 使用 Python 3.10+ 语法
- 遵循 PEP 8 代码风格
- 添加适当的注释和文档字符串

## 🐛 常见问题

### Q: 提示"文件夹不存在"错误
A: 检查配置文件中的 `path` 值是否正确，确保文件夹ID有效

### Q: 代理连接失败
A: 确认代理地址和端口正确，检查代理服务是否正常运行

### Q: RSS 更新延迟
A: Mikan RSS 有一定延迟，请耐心等待，或调整检查间隔时间

### Q: 种子重复添加
A: 程序会自动检测重复内容，如果仍有问题，检查日志文件

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [YinBuLiao](https://github.com/YinBuLiao) - 项目原作者，创建了 [Bangumi-PikPak](https://github.com/YinBuLiao/Bangumi-PikPak) 项目
- [Mikan Project](https://mikanani.me) - 提供番剧 RSS 源
- [PikPak](https://mypikpak.com/) - 提供云存储服务
- [pikpakapi](https://github.com/Quan666/PikPakAPI) - PikPak API 封装

## 📞 联系方式

- 项目主页：https://github.com/hrWong/Bangumi-PikPak
- 问题反馈：https://github.com/hrWong/Bangumi-PikPak/issues
- 邮箱：h.r.wong@foxmail.com

---

如果这个项目对你有帮助，请给个 ⭐ Star 支持一下！
