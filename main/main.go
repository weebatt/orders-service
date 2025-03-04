package main

import (
	"310499-itmobatareyka-course-1343/internal/repository"
	"310499-itmobatareyka-course-1343/internal/service"
	test "310499-itmobatareyka-course-1343/pkg/api/test/proto/api"
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:50051", "gRPC server endpoint")
)

func main() {
	errChan := make(chan error)
	stopChan := make(chan os.Signal)

	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	httpServer := http.Server{Addr: ":8081"}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := test.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)

	fmt.Println("starting http server on 8081")

	go func() {
		if err = httpServer.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	orderRepo := repository.InitializationOrderRepository()
	orderService := service.InitializationOrderService(orderRepo)

	grpcServer := grpc.NewServer()
	test.RegisterOrderServiceServer(grpcServer, orderService)

	listener, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

	fmt.Println("gRPC-сервер запущен на порту 50051")
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	defer func() {
		fmt.Println("Server Stopped")
		_ = httpServer.Shutdown(shutdownCtx)
		grpcServer.GracefulStop()
	}()

	select {
	case err := <-errChan:
		log.Printf("Fatal error: %v\n", err)
	case <-stopChan:
	}
}
