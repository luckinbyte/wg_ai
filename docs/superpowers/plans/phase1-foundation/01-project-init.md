# Task 1: Project Initialization - 项目初始化

## 背景与目标

这是整个 Go 游戏服务器项目的第一个任务。目标是创建基础的 Go 项目结构，为后续开发奠定基础。

**为什么需要这个任务：**
- 建立统一的目录结构，方便团队协作
- 创建 Makefile 统一构建流程
- 初始化 Go module 管理依赖

**输出：**
- `go.mod` - Go 模块定义
- `Makefile` - 构建脚本
- 目录结构骨架

## 依赖

无前置依赖，这是项目的起点。

## 步骤

### Step 1: Initialize Go module

```bash
cd /root/ai_project/wg_ai
go mod init github.com/yourorg/wg_ai
```

Expected output:
```
go: creating new go.mod: module github.com/yourorg/wg_ai
```

### Step 2: Create Makefile

Create file `Makefile`:

```makefile
.PHONY: all build test clean proto

all: proto build

build:
	go build -o bin/game ./cmd/game
	go build -o bin/login ./cmd/login
	go build -o bin/db ./cmd/db

test:
	go test -v ./...

clean:
	rm -rf bin/

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/cs/*.proto proto/ss/*.proto
```

### Step 3: Create directory structure

```bash
cd /root/ai_project/wg_ai
mkdir -p cmd/game cmd/login cmd/db
mkdir -p internal/common/config internal/common/logger internal/common/errors
mkdir -p internal/gate internal/agent internal/session internal/rpc internal/db
mkdir -p proto/cs proto/ss
mkdir -p config
```

### Step 4: Commit

```bash
git add .
git commit -m "chore: initialize project structure"
```

## 验证

```bash
ls -la /root/ai_project/wg_ai/
```

Expected: 应看到 cmd/, internal/, proto/, config/ 目录和 go.mod, Makefile 文件

## 完成标志

- [ ] go.mod 文件存在
- [ ] Makefile 文件存在
- [ ] 目录结构创建完成
- [ ] 首次 commit 完成
