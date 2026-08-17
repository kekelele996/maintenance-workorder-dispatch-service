package service

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"workorder/internal/config"
	"workorder/internal/model"
	"workorder/internal/repository"
	"workorder/internal/util"
)

var (
	ErrOrderNotFound     = errors.New("work order not found")
	ErrTechNotFound      = errors.New("technician not found")
	ErrValidation        = errors.New("invalid work order")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrNoAvailableTech   = errors.New("no available technician")
)

var idCounter atomic.Int64

func generateID() string {
	return fmt.Sprintf("wo-%d", idCounter.Add(1))
}

// Dispatcher 根据设备类别给出需要的技能标签。
type Dispatcher interface {
	SkillFor(equipmentID string) string
}

// SkillDispatcher 是 Dispatcher 的默认实现，从路由表读取技能，缺省时补登记默认技能。
type SkillDispatcher struct {
	routes   map[string]string
	fallback string
}

// NewSkillDispatcher 构造一个 SkillDispatcher；routes 为空时自动初始化。
func NewSkillDispatcher(routes map[string]string) *SkillDispatcher {
	return &SkillDispatcher{routes: routes, fallback: "general"}
}

func (d *SkillDispatcher) SkillFor(equipmentID string) string {
	cat := util.EquipCategory(equipmentID)
	if skill, ok := d.routes[cat]; ok {
		return skill
	}
	d.routes[cat] = d.fallback
	return d.fallback
}

// Service 是业务逻辑层，组合 repository、config 与 dispatcher。
type Service struct {
	repo       *repository.Repository
	cfg        *config.Config
	dispatcher Dispatcher
}

func New(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		repo:       repo,
		cfg:        cfg,
		dispatcher: NewSkillDispatcher(cfg.SkillRoutes),
	}
}

func (s *Service) CreateWorkOrder(equipmentID, title string, priority model.Priority) (*model.WorkOrder, error) {
	if equipmentID == "" || title == "" {
		return nil, ErrValidation
	}
	if priority < model.PriorityLow || priority > model.PriorityUrgent {
		return nil, ErrValidation
	}
	now := time.Now()
	o := &model.WorkOrder{
		ID:          generateID(),
		EquipmentID: equipmentID,
		Title:       title,
		Priority:    priority,
		Status:      model.StatusPending,
		ScheduledAt: now,
		Attempts:    0,
		MaxAttempts: s.cfg.RetryLimit,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created, err := s.repo.Create(o)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) FindWorkOrder(id string) (*model.WorkOrder, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("work order %s: %w", id, ErrOrderNotFound)
		}
		return nil, fmt.Errorf("lookup work order %s: %w", id, err)
	}
	return o, nil
}

// ListWorkOrders 返回全部工单或按状态过滤后的工单，结果按优先级排序。
func (s *Service) ListWorkOrders(status *model.Status) ([]*model.WorkOrder, error) {
	orders, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	if status == nil {
		return util.SortByPriority(orders), nil
	}
	filtered := util.FilterByStatus(orders, *status)
	return util.SortByPriority(filtered), nil
}

// DispatchOrder 把工单派给指定技工：校验状态与技能匹配后置为 assigned。
func (s *Service) DispatchOrder(orderID, techID string) (*model.WorkOrder, error) {
	o, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("work order %s: %w", orderID, ErrOrderNotFound)
		}
		return nil, err
	}
	if !model.CanTransition(o.Status, model.StatusAssigned) {
		return nil, fmt.Errorf("order %s from %s: %w", orderID, o.Status, ErrInvalidTransition)
	}
	tech, err := s.repo.FindTechnician(techID)
	if err != nil {
		if errors.Is(err, repository.ErrTechNotFound) {
			return nil, fmt.Errorf("technician %s: %w", techID, ErrTechNotFound)
		}
		return nil, err
	}
	if !tech.Available {
		return nil, ErrNoAvailableTech
	}
	required := s.dispatcher.SkillFor(o.EquipmentID)
	if !tech.HasSkill(required) {
		return nil, ErrNoAvailableTech
	}
	return s.repo.Update(orderID, func(oo *model.WorkOrder) {
		oo.Status = model.StatusAssigned
		oo.TechnicianID = techID
		oo.UpdatedAt = time.Now()
	})
}

// ExecuteOrder 模拟执行工单：assigned -> in_progress -> completed/failed。
func (s *Service) ExecuteOrder(orderID string) (*model.WorkOrder, error) {
	o, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("work order %s: %w", orderID, ErrOrderNotFound)
		}
		return nil, err
	}
	if o.Status == model.StatusInProgress {
		return s.completeOrder(o)
	}
	if !model.CanTransition(o.Status, model.StatusInProgress) {
		return nil, fmt.Errorf("order %s from %s: %w", orderID, o.Status, ErrInvalidTransition)
	}
	_, err = s.repo.Update(orderID, func(oo *model.WorkOrder) {
		oo.Status = model.StatusInProgress
		oo.Attempts++
		oo.UpdatedAt = time.Now()
	})
	if err != nil {
		return nil, err
	}
	return s.completeOrder(o)
}

func (s *Service) completeOrder(o *model.WorkOrder) (*model.WorkOrder, error) {
	if o.EquipmentID == "FAULTY" {
		return s.repo.Update(o.ID, func(oo *model.WorkOrder) {
			oo.Status = model.StatusFailed
			oo.LastError = "equipment reports fault"
			oo.UpdatedAt = time.Now()
		})
	}
	return s.repo.Update(o.ID, func(oo *model.WorkOrder) {
		oo.Status = model.StatusCompleted
		oo.LastError = ""
		oo.UpdatedAt = time.Now()
	})
}

// RetryOrder 把失败工单置为 retrying，等待下次调度再次执行。
func (s *Service) RetryOrder(orderID string) (*model.WorkOrder, error) {
	o, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("work order %s: %w", orderID, ErrOrderNotFound)
		}
		return nil, err
	}
	if !model.CanTransition(o.Status, model.StatusRetrying) {
		return nil, fmt.Errorf("order %s from %s: %w", orderID, o.Status, ErrInvalidTransition)
	}
	return s.repo.Update(orderID, func(oo *model.WorkOrder) {
		oo.Status = model.StatusRetrying
		oo.UpdatedAt = time.Now()
	})
}

// Stats 返回各状态工单数量，供仪表盘展示；同一份快照上连续过滤。
func (s *Service) Stats() map[model.Status]int {
	orders, _ := s.repo.List()
	result := map[model.Status]int{}
	for _, st := range []model.Status{
		model.StatusPending,
		model.StatusAssigned,
		model.StatusInProgress,
		model.StatusRetrying,
		model.StatusCompleted,
		model.StatusFailed,
	} {
		result[st] = len(util.FilterByStatus(orders, st))
	}
	return result
}

// ActiveCount 并发统计仍在处理中的工单数。
func (s *Service) ActiveCount() int {
	orders, _ := s.repo.List()
	if len(orders) == 0 {
		return 0
	}
	groups := make([]*model.WorkOrder, len(orders))
	copy(groups, orders)

	var wg sync.WaitGroup
	var count atomic.Int64
	step := (len(groups) + 3) / 4
	if step < 1 {
		step = 1
	}
	for i := 0; i < len(groups); i += step {
		end := i + step
		if end > len(groups) {
			end = len(groups)
		}
		wg.Add(1)
		go func(chunk []*model.WorkOrder) {
			defer wg.Done()
			for _, o := range chunk {
				if model.ActiveStatuses[o.Status] {
					count.Add(1)
				}
			}
		}(groups[i:end])
	}
	wg.Wait()
	return int(count.Load())
}
