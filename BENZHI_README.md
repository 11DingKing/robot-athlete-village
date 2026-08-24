# BENZHI_README

这是一个 Go 后端服务，用于机器人“运动员村”运营后端：管理代表团入住、训练场地时段、设备维护、赛事签到和审计。

## 项目说明

- 项目：11DingKing/robot-athlete-village
- 项目用途：机器人“运动员村”运营后端：管理代表团入住、训练场地时段、设备维护、赛事签到和审计。项目为 compact10 基础仓库，采用 Go、SQLite、版本化迁移和 HTTP API。
- Go 工具链：`golang:1.26`
- 前端工具链：无

## 标准构建、运行和测试命令

进入容器后执行：

```bash
# 编译
cd '/app' && GOTOOLCHAIN=local go build ./...

# 启动
cd '/app' && GOTOOLCHAIN=local go run ./cmd/server

# 测试
cd '/app' && GOTOOLCHAIN=local go test ./...
```

## Docker 构建和进入容器

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-task-5-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-task-5-arm64 linux/arm64
docker run -it benzhi-task-5-amd64:latest
docker run -it --platform linux/arm64 benzhi-task-5-arm64:latest
```

## 题目验证命令

1. 预期退出码 0：`go test ./internal/worker -run '^TestMaintenanceCompletionRollsBackWhenEquipmentRestoreFails$' -count=1`
2. 预期退出码 0：`go test ./...`
3. 预期退出码 0：`GOTOOLCHAIN=local go build -buildvcs=false ./... && GOTOOLCHAIN=local go vet ./...`
