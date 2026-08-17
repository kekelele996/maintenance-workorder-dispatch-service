package repository

import (
	"errors"
	"fmt"

	"workorder/internal/model"
	"workorder/internal/store"
)

var (
	ErrNotFound      = errors.New("work order not found")
	ErrAlreadyExists = errors.New("work order already exists")
	ErrTechNotFound  = errors.New("technician not found")
)

// Repository 在内存 Store 之上提供带错误语义的数据访问层。
type Repository struct {
	store *store.Store
}

func New(s *store.Store) *Repository {
	return &Repository{store: s}
}

func (r *Repository) Create(o *model.WorkOrder) (*model.WorkOrder, error) {
	if err := r.store.PutOrder(o); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, fmt.Errorf("create order %s: %w", o.ID, ErrAlreadyExists)
		}
		return nil, fmt.Errorf("create order %s: %w", o.ID, err)
	}
	return o, nil
}

func (r *Repository) FindByID(id string) (*model.WorkOrder, error) {
	o, err := r.store.GetOrder(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("work order %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("work order %s: %w", id, err)
	}
	return o, nil
}

func (r *Repository) FindByEquipment(equipmentID string) ([]*model.WorkOrder, error) {
	orders := r.store.ListOrders()
	out := make([]*model.WorkOrder, 0)
	for _, o := range orders {
		if o.EquipmentID == equipmentID {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *Repository) List() ([]*model.WorkOrder, error) {
	return r.store.ListOrders(), nil
}

func (r *Repository) Update(id string, fn func(*model.WorkOrder)) (*model.WorkOrder, error) {
	o, err := r.store.UpdateOrder(id, fn)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("work order %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("work order %s: %w", id, err)
	}
	return o, nil
}

func (r *Repository) CreateTechnician(t *model.Technician) (*model.Technician, error) {
	if err := r.store.PutTechnician(t); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, fmt.Errorf("technician %s: %w", t.ID, ErrAlreadyExists)
		}
		return nil, fmt.Errorf("technician %s: %w", t.ID, err)
	}
	return t, nil
}

func (r *Repository) FindTechnician(id string) (*model.Technician, error) {
	t, err := r.store.GetTechnician(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("technician %s: %w", id, ErrTechNotFound)
		}
		return nil, fmt.Errorf("technician %s: %w", id, err)
	}
	return t, nil
}

func (r *Repository) ListTechnicians() ([]*model.Technician, error) {
	return r.store.ListTechnicians(), nil
}
