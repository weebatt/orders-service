package cmd

import (
	"310499-itmobatareyka-course-1343/internal/repository"
	"310499-itmobatareyka-course-1343/internal/service"
	test "310499-itmobatareyka-course-1343/pkg/api/test/api"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	orderRepo := repository.InitializationOrderRepository()
	orderService := service.InitializationOrderService(orderRepo)

	grpcServer := grpc.NewServer()
	test.RegisterOrderServiceServer(grpcServer, orderService)

	listener, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

	fmt.Println("gRPC-сервер запущен на порту 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка при запуске gRPC сервера: %v", err)
	}
}
