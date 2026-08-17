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

// PutOrder 写入工单；存储内部保存入参的副本，斩断与调用方的指针共享。
func (s *Store) PutOrder(o *model.WorkOrder) error {
	if o == nil || o.ID == "" {
		return errors.New("invalid work order")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orders[o.ID]; ok {
		return ErrAlreadyExists
	}
	s.orders[o.ID] = o.Clone()
	s.orderIDs = append(s.orderIDs, o.ID)
	return nil
}

// GetOrder 返回工单的深拷贝，调用方修改不影响存储内对象。
func (s *Store) GetOrder(id string) (*model.WorkOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	return o.Clone(), nil
}

// ListOrders 返回按插入顺序排列的工单副本切片，元素互不影响。
func (s *Store) ListOrders() []*model.WorkOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.WorkOrder, 0, len(s.orderIDs))
	for _, id := range s.orderIDs {
		out = append(out, s.orders[id].Clone())
	}
	return out
}

// OrderIDs 返回按插入顺序排列的工单 ID 副本。
func (s *Store) OrderIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.orderIDs...)
}

// UpdateOrder 在锁内对工单执行原地修改，返回修改后的深拷贝。
// 返回副本而非内部引用，避免调用方在锁外读写存储对象引发 data race。
func (s *Store) UpdateOrder(id string, fn func(*model.WorkOrder)) (*model.WorkOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	fn(o)
	return o.Clone(), nil
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.orderIDs)
}

// PutTechnician 写入技工；存储内部保存入参的副本，Skills 使用独立底层数组。
func (s *Store) PutTechnician(t *model.Technician) error {
	if t == nil || t.ID == "" {
		return errors.New("invalid technician")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.technicians[t.ID]; ok {
		return ErrAlreadyExists
	}
	s.technicians[t.ID] = t.Clone()
	s.techIDs = append(s.techIDs, t.ID)
	return nil
}

// GetTechnician 返回技工的深拷贝，Skills 切片使用独立底层数组。
func (s *Store) GetTechnician(id string) (*model.Technician, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.technicians[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t.Clone(), nil
}

// ListTechnicians 返回按插入顺序排列的技工副本切片。
func (s *Store) ListTechnicians() []*model.Technician {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Technician, 0, len(s.techIDs))
	for _, id := range s.techIDs {
		out = append(out, s.technicians[id].Clone())
	}
	return out
}

// TechnicianIDs 返回按插入顺序排列的技工 ID 副本。
func (s *Store) TechnicianIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.techIDs...)
}
