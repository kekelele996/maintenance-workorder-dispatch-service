# BUG_REPRO

## Bug 是什么

自定义技能路由配置在解析结果为空时返回 nil map，dispatcher 构造时没有兜底初始化，后续按设备类别写路由时向 nil map 写入触发 panic。

## 如何触发

```bash
go test -run TestSkillRouteFallback ./...
```

## 错误信息

```
--- FAIL: TestSkillRouteFallback
panic: assignment to entry in nil map
```
