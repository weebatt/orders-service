package service

import (
	"310499-itmobatareyka-course-1343/internal/repository"
	test "310499-itmobatareyka-course-1343/pkg/api/test/proto/api"
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServiceServer struct {
	test.UnimplementedOrderServiceServer
	orderRepository *repository.OrderMap
}

func InitializationOrderService(orderMap *repository.OrderMap) *OrderServiceServer {
	return &OrderServiceServer{orderRepository: orderMap}
}

func (os *OrderServiceServer) CreateOrder(ctx context.Context, in *test.CreateOrderRequest) (*test.CreateOrderResponse, error) {
	if in.Item == "" {
		return nil, status.Error(codes.InvalidArgument, "empty item")
	} else if in.Quantity == 0 {
		return nil, status.Error(codes.InvalidArgument, "zero quantity")
	}

	id := uuid.New().String()
	order := test.Order{Id: id, Item: in.Item, Quantity: in.Quantity}
	_ = os.orderRepository.Create(&order)

	return &test.CreateOrderResponse{Id: id}, nil
}

func (os *OrderServiceServer) GetOrder(ctx context.Context, in *test.GetOrderRequest) (*test.GetOrderResponse, error) {
	id := in.Id

	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	order, _ := os.orderRepository.Find(id)

	return &test.GetOrderResponse{Order: order}, nil
}

func (os *OrderServiceServer) UpdateOrder(ctx context.Context, in *test.UpdateOrderRequest) (*test.UpdateOrderResponse, error) {
	id := in.Id

	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	} else if in.Quantity == 0 {
		return nil, status.Error(codes.InvalidArgument, "zero quantity")
	} else if in.Item == "" {
		return nil, status.Error(codes.InvalidArgument, "empty item")
	}

	order := test.Order{Id: id, Item: in.Item, Quantity: in.Quantity}
	update, _ := os.orderRepository.Update(&order)

	return &test.UpdateOrderResponse{Order: update}, nil
}

func (os *OrderServiceServer) DeleteOrder(ctx context.Context, in *test.DeleteOrderRequest) (*test.DeleteOrderResponse, error) {
	if in.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	success, _ := os.orderRepository.Delete(in.Id)

	return &test.DeleteOrderResponse{Success: success}, nil
}

func (os *OrderServiceServer) ListOrders(ctx context.Context, in *test.ListOrdersRequest) (*test.ListOrdersResponse, error) {
	orders, _ := os.orderRepository.GetAllOrders()
	return &test.ListOrdersResponse{Orders: orders}, nil
}
