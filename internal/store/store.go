package store

import (
	"errors"
	"sync"

	"workorder/internal/model"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// Store 是内存存储，用读写锁保护所有工单与技工数据。
type Store struct {
	mu          sync.RWMutex
	orders      map[string]*model.WorkOrder
	orderIDs    []string
	technicians map[string]*model.Technician
	techIDs     []string
}

func New() *Store {
	return &Store{
		orders:      make(map[string]*model.WorkOrder),
		orderIDs:    []string{},
		technicians: make(map[string]*model.Technician),
		techIDs:     []string{},
	}
}

func cloneOrder(o *model.WorkOrder) *model.WorkOrder {
	if o == nil {
		return nil
	}
	c := *o
	return &c
}

func cloneTechnician(t *model.Technician) *model.Technician {
	if t == nil {
		return nil
	}
	c := *t
	c.Skills = append([]string(nil), t.Skills...)
	return &c
}

func (s *Store) PutOrder(o *model.WorkOrder) error {
	if o == nil || o.ID == "" {
		return errors.New("invalid work order")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orders[o.ID]; ok {
		return ErrAlreadyExists
	}
	s.orders[o.ID] = cloneOrder(o)
	s.orderIDs = append(s.orderIDs, o.ID)
	return nil
}

func (s *Store) GetOrder(id string) (*model.WorkOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneOrder(o), nil
}

func (s *Store) ListOrders() []*model.WorkOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.WorkOrder, 0, len(s.orderIDs))
	for _, id := range s.orderIDs {
		out = append(out, cloneOrder(s.orders[id]))
	}
	return out
}

func (s *Store) OrderIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.orderIDs))
	copy(out, s.orderIDs)
	return out
}

// UpdateOrder 在锁内对工单执行原地修改，返回值是修改后的副本。
func (s *Store) UpdateOrder(id string, fn func(*model.WorkOrder)) (*model.WorkOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	fn(o)
	return cloneOrder(o), nil
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.orderIDs)
}

func (s *Store) PutTechnician(t *model.Technician) error {
	if t == nil || t.ID == "" {
		return errors.New("invalid technician")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.technicians[t.ID]; ok {
		return ErrAlreadyExists
	}
	s.technicians[t.ID] = cloneTechnician(t)
	s.techIDs = append(s.techIDs, t.ID)
	return nil
}

func (s *Store) GetTechnician(id string) (*model.Technician, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.technicians[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneTechnician(t), nil
}

func (s *Store) ListTechnicians() []*model.Technician {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Technician, 0, len(s.techIDs))
	for _, id := range s.techIDs {
		out = append(out, cloneTechnician(s.technicians[id]))
	}
	return out
}

func (s *Store) TechnicianIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.techIDs))
	copy(out, s.techIDs)
	return out
}
