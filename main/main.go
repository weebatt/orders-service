package main

import (
	"310499-itmobatareyka-course-1343/internal/config"
	"310499-itmobatareyka-course-1343/internal/repository"
	"310499-itmobatareyka-course-1343/internal/service"
	test "310499-itmobatareyka-course-1343/pkg/api/test/proto/api"
	"310499-itmobatareyka-course-1343/pkg/logger"
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	_ "go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// def chains for graceful shutdown
	errChan := make(chan error)
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	// def context and config
	ctx := context.Background()
	ctx, _ = logger.New(ctx)
	cfg, err := config.New()

	if err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed reading config", zap.Error(err))
	}

	// def http server
	httpServer := http.Server{Addr: ":" + strconv.Itoa(cfg.HTTPPort)}
	grpcServerEndpoint := flag.String("grpc-server-endpoint", cfg.GRPCServerEndpoint, "gRPC server endpoint")

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	errors := test.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)

	if errors != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to register grpc server", zap.Error(errors))
	}

	// def grpc server
	orderRepo := repository.InitializationOrderRepository()
	orderService := service.InitializationOrderService(orderRepo)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor))
	test.RegisterOrderServiceServer(grpcServer, orderService)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.HTTPPort))

	if err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to listen", zap.Error(err))
	}

	// goroutines for checking SIGTERM and SIGINT
	logger.GetLoggerFromContext(ctx).Info(ctx, "http server successfully started", zap.Int("port", cfg.HTTPPort))
	go func() {
		if errors = httpServer.ListenAndServe(); errors != nil {
			errChan <- errors
		}
	}()

	logger.GetLoggerFromContext(ctx).Info(ctx, "grpc server successfully started", zap.Int("port", cfg.GRPCPort))
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	// stopping servers
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	defer func() {
		fmt.Println("Server Stopped")
		_ = httpServer.Shutdown(shutdownCtx)
		grpcServer.GracefulStop()
	}()

	select {
	case err := <-errChan:
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to serve", zap.Error(err))
	case <-stopChan:
	}
}
