# BUG_REPRO

## Bug 是什么

过滤工具用 orders[:0] 原地复用底层数组，污染了调用方持有的同一份工单快照；统计方法在同一份快照上连续按状态过滤，读到被污染的切片，返回的状态数量不对。

## 如何触发

```bash
go test ./...
```

## 错误信息

```
--- FAIL: TestFilterByStatusNoAliasing
    util_test.go:25: FilterByStatus corrupted input order
--- FAIL: TestStats
    service_test.go:161: pending=1 want 2
```
