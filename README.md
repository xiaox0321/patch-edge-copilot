# Patch Edge Copilot

一个用于自动启用Microsoft Edge浏览器Copilot功能的跨平台工具。

## 致谢

本项目基于 [jiarandiana0307/patch-edge-copilot](https://github.com/jiarandiana0307/patch-edge-copilot) 的Python版本移植而来。

感谢原作者 [jiarandiana0307](https://github.com/jiarandiana0307) 提供的核心实现逻辑和思路。本项目在此基础上：
- 迁移到Go语言以实现跨平台编译
- 添加GitHub Actions自动化构建
- 提供多平台预编译二进制文件

## 功能

- 自动检测Microsoft Edge浏览器安装路径（支持Stable、Beta、Dev、Canary版本）
- 安全关闭正在运行的Edge进程
- 自动修改Edge配置以启用Copilot功能
  - 设置`variations_country`为`US`
  - 启用`chat_ip_eligibility_status`
- 支持Windows、Linux、macOS平台
- 支持x64和ARM64架构

## 使用方法

### 手动编译

```bash
# 克隆仓库
git clone https://github.com/xiaox0321/patch-edge-copilot.git
cd patch-edge-copilot

# 编译（会自动下载依赖）
go mod download
go build -o patch-edge-copilot main.go
```

### 交叉编译

```bash
# 编译为Windows (x64)
GOOS=windows GOARCH=amd64 go build -o patch-edge-copilot.exe main.go

# 编译为Linux (ARM64)
GOOS=linux GOARCH=arm64 go build -o patch-edge-copilot main.go

# 编译为macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o patch-edge-copilot main.go
```

### 运行

```bash
# Linux/macOS
./patch-edge-copilot

# Windows
patch-edge-copilot.exe
```

## 自动化构建

项目包含GitHub Actions工作流，支持自动构建和发布：

### Release流程

1. 推送版本标签：
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub Actions将自动构建所有平台的二进制文件

3. 创建包含所有下载链接的Release

### 支持的平台和架构

- **Windows**: x64, ARM64
- **Linux**: x64, ARM64
- **macOS**: Intel (x64), Apple Silicon (arm64)

## 工作原理

该工具通过以下方式启用Edge Copilot：

1. 定位Edge用户数据目录
2. 关闭所有Edge进程（避免文件锁定）
3. 修改`Local State`文件：
   - 设置`variations_country`为`US`
4. 修改每个Profile的`Preferences`文件：
   - 设置`browser.chat_ip_eligibility_status`为`true`
5. 重启Edge浏览器

## 注意事项

- 需要管理员/root权限来修改Edge配置文件
- 运行时需要关闭Edge浏览器
- 修改的文件位于Edge用户数据目录，请备份重要数据
- macOS上可能需要在"系统偏好设置 > 安全性与隐私"中允许运行

## 许可证

MIT License
