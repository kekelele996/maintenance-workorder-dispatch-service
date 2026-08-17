# BUG_REPRO

## Bug 是什么
存储层读路径去掉了深拷贝，GetOrder / ListOrders / GetTechnician / ListTechnicians / OrderIDs / TechnicianIDs 直接返回内部引用；service 的 ActiveCount 直接并发聚合这些引用；worker 执行阶段用 goroutine 并发执行并在 goroutine 内 Add WaitGroup、同时累加共享计数器。

## 如何触发
`cd <env> && go test ./... -race -count=20`

## 错误信息
- store 单测报“返回内部引用被改坏”（TestGetOrderReturnsCopy / TestListOrdersReturnsCopies / TestTechnicians）。
- `-race` 在 service.TestConcurrentDispatchAndRead 与 worker.TestTickRetriesAndExecutes 报 DATA RACE（并发读写、WaitGroup 计数错配、共享计数）。
