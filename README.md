# Robot Athlete Village

机器人“运动员村”运营后端：管理代表团入住、训练场地时段、设备维护、赛事签到和审计。项目为 compact_10 基础仓库，采用 Go、SQLite、版本化迁移和 HTTP API。

## 运行

```bash
go run ./cmd/server
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

默认账号：`mayor@example.com` / `village-secret`（village_admin），`coach@example.com` / `coach-secret`（coach）。

## 业务边界

入住申请必须通过审核后才能分配房间；训练预约要校验场地时段和教练资格；设备维护会生成可重试的后台作业；赛事签到写入审计和幂等事件。所有请求携带 request id，数据库迁移可重复执行。
