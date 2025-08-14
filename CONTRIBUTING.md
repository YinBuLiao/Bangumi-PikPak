# 贡献指南

感谢您对 Bangumi-PikPak 项目的关注！我们欢迎所有形式的贡献。

## 如何贡献

### 报告 Bug

如果您发现了 bug，请：

1. 检查现有的 [Issues](https://github.com/hrWong/Bangumi-PikPak/issues) 是否已经报告
2. 如果没有，请创建一个新的 Issue
3. 使用我们提供的 bug 报告模板
4. 提供详细的错误信息和重现步骤

### 功能请求

如果您有功能建议，请：

1. 检查现有的 [Issues](https://github.com/hrWong/Bangumi-PikPak/issues) 是否已经提出
2. 如果没有，请创建一个新的 Issue
3. 详细描述您想要的功能和用例

### 代码贡献

如果您想贡献代码，请：

1. Fork 这个仓库
2. 创建一个新的分支：`git checkout -b feature/proxy-support`
3. 进行您的更改
4. 提交您的更改：`git commit -m 'Add some feature'`
5. 推送到分支：`git push origin feature/proxy-support`
6. 创建一个 Pull Request

## 开发环境设置

### 环境要求

- Python 3.10+
- Git

### 设置步骤

1. **Fork 并克隆仓库**
```bash
git clone https://github.com/hrWong/Bangumi-PikPak.git
cd Bangumi-PikPak
```

2. **创建虚拟环境**
```bash
python -m venv venv
source venv/bin/activate  # Linux/macOS
# 或
venv\Scripts\activate     # Windows
```

3. **安装依赖**
```bash
pip install -r requirements.txt
```

4. **安装开发依赖**
```bash
pip install -r requirements-dev.txt  # 如果有的话
```

## 提交规范

### 提交消息格式

```
<类型>(<范围>): <描述>

[可选的详细描述]

[可选的脚注]
```

### 类型

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### 示例

```
feat(proxy): 添加SOCKS代理支持

- 支持SOCKS5代理协议
- 自动检测代理类型
- 添加代理测试功能

Closes #123
```

## 审查流程

1. 所有代码更改都需要通过 Pull Request
2. 至少需要一名维护者审查并批准
3. 代码审查通过后，维护者会合并代码
4. 如果有问题，维护者会要求修改

## 问题反馈

如果您在贡献过程中遇到任何问题，请：

1. 查看 [Issues](https://github.com/YinBuLiao/Bangumi-PikPak/issues)
2. 创建新的 Issue 描述您的问题

## 行为准则

我们致力于为每个人提供友好、安全和欢迎的环境。请：

- 尊重所有贡献者
- 保持专业和礼貌
- 接受建设性的批评
- 专注于问题本身，而不是个人

## 许可证

通过贡献代码，您同意您的贡献将在 MIT 许可证下发布。

## 感谢

再次感谢您对 Bangumi-PikPak 项目的贡献！您的参与让这个项目变得更好。
