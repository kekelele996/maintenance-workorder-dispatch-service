# BUG_REPRO

## Bug 是什么

工单重试状态 retrying 没有在状态转换表、重试写回、调度查询、活跃过滤四处同步：失败单切不到 retrying，重试单也不会被调度执行。

## 如何触发

```bash
go test ./...
```

## 错误信息

```
--- FAIL: TestCanTransition
    model_test.go:21: CanTransition(retrying,in_progress)=false want true
--- FAIL: TestRetryLifecycle
    service_test.go:103: status="failed" want retrying
--- FAIL: TestFilterActive
    util_test.go:37: len=1 want 2
--- FAIL: TestTickRetriesAndExecutes
    worker_test.go:36: executed=0 want 1
```
