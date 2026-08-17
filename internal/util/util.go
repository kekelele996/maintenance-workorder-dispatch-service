package util

import (
	"sort"

	"workorder/internal/model"
)

// FilterByStatus 返回状态等于 status 的工单子集，结果是新切片，不影响入参。
func FilterByStatus(orders []*model.WorkOrder, status model.Status) []*model.WorkOrder {
	out := make([]*model.WorkOrder, 0, len(orders))
	for _, o := range orders {
		if o.Status == status {
			out = append(out, o)
		}
	}
	return out
}

// FilterActive 返回仍在处理中（未完成、未失败）的工单子集，结果是新切片。
func FilterActive(orders []*model.WorkOrder) []*model.WorkOrder {
	out := make([]*model.WorkOrder, 0, len(orders))
	for _, o := range orders {
		if o.Status == model.StatusPending || o.Status == model.StatusAssigned || o.Status == model.StatusInProgress {
			out = append(out, o)
		}
	}
	return out
}

// SortByPriority 按优先级从高到低排序，返回新切片，不改变入参顺序。
func SortByPriority(orders []*model.WorkOrder) []*model.WorkOrder {
	out := make([]*model.WorkOrder, len(orders))
	copy(out, orders)
	sort.SliceStable(out, func(i, j int) bool {
		return model.PriorityRank(out[i].Priority) > model.PriorityRank(out[j].Priority)
	})
	return out
}

// TopByPriority 返回优先级最高的 n 张工单，结果是新切片。
func TopByPriority(orders []*model.WorkOrder, n int) []*model.WorkOrder {
	if n <= 0 {
		return []*model.WorkOrder{}
	}
	sorted := SortByPriority(orders)
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

// EquipCategory 从设备编号推导设备类别，供派单路由使用。
func EquipCategory(equipmentID string) string {
	if len(equipmentID) >= 4 && equipmentID[:3] == "CNC" {
		return "cnc"
	}
	if len(equipmentID) >= 4 && equipmentID[:3] == "BLR" {
		return "boiler"
	}
	if len(equipmentID) >= 4 && equipmentID[:3] == "CVY" {
		return "conveyor"
	}
	if len(equipmentID) >= 4 && equipmentID[:3] == "HVC" {
		return "hvac"
	}
	return "default"
}
