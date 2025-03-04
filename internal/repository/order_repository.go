package repository

import (
	test "310499-itmobatareyka-course-1343/pkg/api/test/proto/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

type Repository interface {
	InitializationOrderRepository() *OrderMap
	Create(order *test.Order) error
	Find(id string) (*test.Order, error)
	Update(id string) (*test.Order, error)
	Delete(id string) (bool, error)
	GetAllOrders() ([]*test.Order, error)
}

type OrderMap struct {
	orders map[string]*test.Order
	mutex  sync.RWMutex
}

func InitializationOrderRepository() *OrderMap {
	return &OrderMap{orders: make(map[string]*test.Order)}
}

func (om *OrderMap) Create(order *test.Order) error {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	if _, exists := om.orders[order.Id]; exists {
		return status.Error(codes.AlreadyExists, "order already exists: Create")
	}
	om.orders[order.Id] = order
	return nil
}

func (om *OrderMap) Find(id string) (*test.Order, error) {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	order, exists := om.orders[id]
	if !exists {
		return nil, status.Error(codes.NotFound, "order not found: Find")
	}
	return order, nil
}

func (om *OrderMap) Update(order *test.Order) (*test.Order, error) {
	om.mutex.Lock()
	defer om.mutex.Unlock()
	_, exists := om.orders[order.Id]
	if !exists {
		return nil, status.Error(codes.NotFound, "order not found: Update")
	}
	om.orders[order.Id] = order
	return order, nil
}

func (om *OrderMap) Delete(id string) (bool, error) {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	if om.orders[id] == nil {
		return false, status.Error(codes.NotFound, "order not found: Delete")
	}
	delete(om.orders, id)
	return true, nil
}

func (om *OrderMap) GetAllOrders() ([]*test.Order, error) {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	orders := make([]*test.Order, 0, len(om.orders))

	for _, order := range om.orders {
		orders = append(orders, order)
	}

	return orders, nil
}
