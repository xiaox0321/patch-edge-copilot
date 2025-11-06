# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个用于自动启用Microsoft Edge浏览器Copilot功能的跨平台工具。该项目将原有的Python实现（[jiarandiana0307/patch-edge-copilot](https://github.com/jiarandiana0307/patch-edge-copilot)）迁移到了Go语言，以便通过GitHub Actions自动构建多平台可执行文件。

**致谢**：本项目基于原作者 [jiarandiana0307](https://github.com/jiarandiana0307) 的Python实现，感谢其提供的核心逻辑和思路。

## 核心功能

1. **路径检测**: 自动检测Microsoft Edge用户数据路径，支持Stable、Beta、Dev、Canary版本
2. **进程管理**: 安全关闭正在运行的Edge进程，避免文件锁定
3. **配置修改**: 自动修改Edge配置启用Copilot功能
   - 设置`variations_country`为`US`
   - 设置`browser.chat_ip_eligibility_status`为`true`
4. **跨平台**: 支持Windows、Linux、macOS，x64和ARM64架构

## 常用命令

### 开发与构建

```bash
# 下载依赖
go mod download

# 本地编译
go build -o patch-edge-copilot main.go

# 交叉编译为不同平台
GOOS=windows GOARCH=amd64 go build -o patch-edge-copilot.exe main.go
GOOS=linux GOARCH=arm64 go build -o patch-edge-copilot main.go
GOOS=darwin GOARCH=arm64 go build -o patch-edge-copilot main.go

# 运行
./patch-edge-copilot
```

### GitHub Actions自动构建

推送到版本标签时自动触发构建：
```bash
git tag v1.0.0
git push origin v1.0.0
```

这将构建所有平台的二进制文件并创建Release。

## 代码架构

### 主要文件结构

- `main.go`: 核心逻辑，包含所有功能实现
- `go.mod`: Go模块定义和依赖管理
- `.github/workflows/build.yml`: GitHub Actions构建工作流
- `README.md`: 项目文档和使用说明

### 关键函数（main.go）

1. **`getVersionAndUserDataPath()`**: 检测Edge安装路径
   - 支持Windows/Linux/macOS
   - 返回存在的版本（stable/canary/dev/beta）

2. **`shutdownEdge()`**: 关闭Edge进程
   - 使用`gopsutil`库枚举进程
   - 避免杀死父进程同名的子进程
   - 返回被终止的进程路径列表

3. **`patchLocalState()`**: 修改Local State文件
   - 将`variations_country`设置为`US`
   - 备份并格式化JSON文件

4. **`patchPreferences()`**: 修改Preferences文件
   - 遍历所有用户配置文件（Default, Profile 1, etc.）
   - 将`browser.chat_ip_eligibility_status`设置为`true`

5. **`getLastVersion()`**: 读取Last Version文件
   - 获取Edge版本号信息

### 依赖库

- **github.com/shirou/gopsutil/v3**: 跨平台进程管理库
  - 用于枚举和终止Edge进程
  - 支持Windows/Linux/macOS

## 开发注意事项

### 权限要求
- 需要管理员/root权限修改Edge配置
- 运行时必须关闭Edge浏览器

### 平台特定行为
- **Windows**: 查找`msedge.exe`进程
- **macOS**: 查找`Microsoft Edge`进程
- **Linux**: 查找`msedge`进程

### 数据安全
- 修改的文件位于用户数据目录
- 建议在修改前备份重要数据

## CI/CD流程

GitHub Actions工作流(`.github/workflows/build.yml`)：
1. 多平台矩阵构建（6种组合）
2. 缓存Go模块加速构建
3. 创建压缩归档
4. 自动发布到GitHub Releases

## 故障排除

### 编译错误
- 确保Go版本≥1.21
- 运行`go mod tidy`清理依赖

### 运行时错误
- 检查Edge是否完全关闭
- 验证用户数据目录权限
- 确保以管理员权限运行
