# glmusage

GLM Coding Plan 用量查询工具。一行命令查看 MCP 额度、Token 限流和今日用量。

## 安装

```bash
# macOS / Linux（自动识别系统和架构）
curl -sL https://github.com/hryyan/glmusage/releases/latest/download/glmusage-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') -o glmusage && chmod +x glmusage && sudo mv glmusage /usr/local/bin/
```

## 使用

```bash
# 设置 Token
export GLM_AUTH_TOKEN="your-token-here"

# 查询用量
glmusage

# 持续监控（每 60 秒刷新）
glmusage -watch

# 自定义刷新间隔
glmusage -watch -interval 30

# 升级到最新版本
glmusage -upgrade
```

## 输出示例

```
GLM [pro] · MCP ██████░░░░░░8% (85/1000) · Token5h ██░░░░░░░░6% · 今日 1,234次 567.9K tok
```
