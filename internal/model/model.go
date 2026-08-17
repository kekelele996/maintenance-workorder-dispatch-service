package model

import "time"

// Priority 表示工单优先级，数值越大越紧急。
type Priority int

const (
	PriorityLow Priority = iota + 1
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

// Status 表示工单状态。
type Status string

const (
	StatusPending    Status = "pending"
	StatusAssigned   Status = "assigned"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusRetrying   Status = "retrying"
)

// ActiveStatuses 是「还在处理中」的状态集合，查询与统计会用到。
var ActiveStatuses = map[Status]bool{
	StatusPending:    true,
	StatusAssigned:   true,
	StatusInProgress: true,
	StatusRetrying:   true,
}

// WorkOrder 是一张设备维保工单。
type WorkOrder struct {
	ID           string
	EquipmentID  string
	Title        string
	Priority     Priority
	Status       Status
	TechnicianID string
	ScheduledAt  time.Time
	Attempts     int
	MaxAttempts  int
	LastError    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Technician 是一名维保技工。
type Technician struct {
	ID        string
	Name      string
	Skills    []string
	Available bool
}

// transitions 是合法状态迁移表。
var transitions = map[Status][]Status{
	StatusPending:    {StatusAssigned, StatusFailed},
	StatusAssigned:   {StatusInProgress, StatusFailed},
	StatusInProgress: {StatusCompleted, StatusFailed},
	StatusFailed:     {StatusRetrying},
	StatusRetrying:   {StatusInProgress},
	StatusCompleted:  {},
}

// CanTransition 判断 from 能否直接迁移到 to。
func CanTransition(from, to Status) bool {
	for _, t := range transitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// PriorityRank 返回优先级对应的整数权重，用于排序。
func PriorityRank(p Priority) int {
	return int(p)
}

// HasSkill 判断技工是否具备指定技能标签。
func (t *Technician) HasSkill(skill string) bool {
	for _, s := range t.Skills {
		if s == skill {
			return true
		}
	}
	return false
}

// Clone 返回工单的深拷贝。调用方拿到副本后任意修改都不会影响存储内的对象，
// 避免跨层共享指针导致的 data race 与内部状态被改写。
func (o *WorkOrder) Clone() *WorkOrder {
	if o == nil {
		return nil
	}
	c := *o
	return &c
}

// Clone 返回技工的深拷贝，Skills 切片使用独立底层数组。
func (t *Technician) Clone() *Technician {
	if t == nil {
		return nil
	}
	c := *t
	if t.Skills != nil {
		c.Skills = append([]string(nil), t.Skills...)
	}
	return &c
}
