package worker

import (
	"context"
	"errors"
	"time"

	"workorder/internal/model"
	"workorder/internal/repository"
	"workorder/internal/service"
)

// Executor 是调度器依赖的工单执行/重试能力，由 service 层实现。
type Executor interface {
	RetryOrder(orderID string) (*model.WorkOrder, error)
	ExecuteOrder(orderID string) (*model.WorkOrder, error)
}

// Scheduler 是后台调度器：周期性捞取失败工单重试、执行重试中的工单。
type Scheduler struct {
	repo         *repository.Repository
	exec         Executor
	pollInterval time.Duration
}

func New(repo *repository.Repository, exec Executor, pollInterval time.Duration) *Scheduler {
	if pollInterval <= 0 {
		pollInterval = 500 * time.Millisecond
	}
	return &Scheduler{repo: repo, exec: exec, pollInterval: pollInterval}
}

// Tick 执行一轮调度：先把失败且未超限的工单置为 retrying，再重新捞取 retrying 工单执行。
// 第二轮重新读取快照，因为第一轮的 RetryOrder 会改写状态，旧快照里的副本不会随之更新。
// 返回本轮重试与执行成功的数量。
func (sch *Scheduler) Tick(ctx context.Context) (retried, executed int) {
	if ctx.Err() != nil {
		return 0, 0
	}

	orders, err := sch.repo.List()
	if err != nil {
		return 0, 0
	}
	for _, o := range orders {
		if o.Status == model.StatusFailed && o.Attempts < o.MaxAttempts {
			_, err := sch.exec.RetryOrder(o.ID)
			if err == nil {
				retried++
			} else if errors.Is(err, service.ErrOrderNotFound) {
				continue
			}
		}
	}

	// 第一轮可能改写了工单状态，重新拉取快照后再执行 retrying 工单。
	orders, err = sch.repo.List()
	if err != nil {
		return retried, 0
	}
	for _, o := range orders {
		if o.Status == model.StatusRetrying {
			if _, err := sch.exec.ExecuteOrder(o.ID); err == nil {
				executed++
			}
		}
	}
	return retried, executed
}

// Run 周期执行 Tick，直到 ctx 被取消。
func (sch *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(sch.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sch.Tick(ctx)
		}
	}
}
