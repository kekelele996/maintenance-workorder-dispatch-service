# BUG_REPRO

## Bug 是什么

数据访问层的哨兵错误在包装时用 %v 丢掉了 %w 错误链，上层 errors.Is 判空失效，导致「工单不存在」被当成系统错误，HTTP 接口把本应 404 的请求映射成 500。

## 如何触发

```bash
go test ./...
```

## 错误信息

```
--- FAIL: TestGetMissingOrderReturns404
    handler_test.go:24: status=500 want 404
--- FAIL: TestFindByIDWrapsNotFound
    repository_test.go:19: errors.Is(err, ErrNotFound)=false
```
