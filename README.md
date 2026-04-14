# glmusage

GLM Coding Plan 用量查询工具。一行命令查看 MCP 额度、Token 限流和今日用量。

## 安装

**使用 Go Install**

```bash
go install github.com/hryyan/glmusage/cmd/glmusage@latest
```

**从源码构建**

```bash
git clone https://github.com/hryyan/glmusage.git
cd glmusage
go build -ldflags "-X main.version=$(git describe --tags --always)" -o glmusage .
```

**从 GitHub Release 下载**

到 [Releases](https://github.com/hryyan/glmusage/releases) 页面下载对应平台的二进制文件。

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

# 查看当前版本
glmusage -version
```

## 输出示例

```
GLM [pro] · MCP ██████░░░░░░8% (85/1000) · Token5h ██░░░░░░░░6% · 今日 1,234次 567.9K tok
```
