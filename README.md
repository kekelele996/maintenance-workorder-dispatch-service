# maintenance-workorder-dispatch-service

一个用 Go 写的设备维保工单调度服务，演示 `handler → service → repository → store → model` 分层、工单状态机与后台调度 worker 的常见写法。

## 功能

- 创建设备维保工单，按优先级与设备类型派单给空闲技工
- 工单状态流转：pending → assigned → in_progress → completed / failed → retrying
- 后台调度器定时捞取到期工单执行，失败自动重试
- 内存存储，读写锁保护，支持按状态 / 技工 / 设备查询

## 目录结构

```
cmd/workorder/        程序入口
internal/config/      环境变量配置
internal/model/       模型定义与状态机
internal/store/       内存存储（工单 / 技工 map）
internal/repository/  数据访问层（错误包装）
internal/service/     业务逻辑（建单 / 派单 / 执行 / 查询）
internal/worker/      后台调度 worker（重试）
internal/handler/     HTTP 接口
internal/util/        过滤 / 排序等工具函数
```

## 运行与测试

```bash
go build ./...          # 编译
go test ./...           # 全量测试
go run ./cmd/workorder  # 启动 HTTP 服务
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `WORKORDER_WORKERS` | 调度 worker 数量 | `4` |
| `WORKORDER_RETRY_LIMIT` | 工单失败重试上限 | `3` |
| `WORKORDER_EXEC_TIMEOUT_MS` | 单次执行超时（毫秒） | `2000` |
| `WORKORDER_POLL_INTERVAL_MS` | 调度轮询间隔（毫秒） | `500` |

## 技术栈

- Go 1.22
