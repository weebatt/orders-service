package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"310499-itmobatareyka-course-1343/internal/config"
	"310499-itmobatareyka-course-1343/internal/repository"
	"310499-itmobatareyka-course-1343/internal/service"
	test "310499-itmobatareyka-course-1343/pkg/api/test/proto/api"
	"310499-itmobatareyka-course-1343/pkg/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// graceful shutdown
	errChan := make(chan error)
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	ctx := context.Background()
	ctx, _ = logger.New(ctx)

	cfg, err := config.New()
	if err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed reading config", zap.Error(err))
	}

	// Подключаемся к PostgreSQL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPass,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to open postgres connection", zap.Error(err))
	}
	defer db.Close()

	// Проверяем, что БД действительно работает
	if err = db.Ping(); err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to ping postgres", zap.Error(err))
	}

	// Инициализируем схему (CREATE TABLE IF NOT EXISTS и т.п.)
	postgresRepo := repository.NewPostgresRepository(db)
	if err := postgresRepo.InitSchema(ctx); err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to init schema", zap.Error(err))
	}

	// готовим gRPC GW
	httpServer := http.Server{Addr: ":" + strconv.Itoa(cfg.HTTPPort)}
	grpcServerEndpoint := flag.String("grpc-server-endpoint", cfg.GRPCServerEndpoint, "gRPC server endpoint")

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := test.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts); err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to register grpc server", zap.Error(err))
	}

	// поднимаем gRPC
	orderService := service.InitializationOrderService(postgresRepo) // <--- наша новая реализация
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor))
	test.RegisterOrderServiceServer(grpcServer, orderService)

	// слушаем gRPC на порту cfg.GRPCPort
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to listen", zap.Error(err))
	}

	// запускаем httpServer (HTTP/REST => gRPC Gateway) на другом порту
	go func() {
		logger.GetLoggerFromContext(ctx).Info(ctx, "http server started", zap.Int("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// запускаем gRPC-сервер
	go func() {
		logger.GetLoggerFromContext(ctx).Info(ctx, "grpc server started", zap.Int("port", cfg.GRPCPort))
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	defer func() {
		fmt.Println("Server Stopped")
		_ = httpServer.Shutdown(shutdownCtx)
		grpcServer.GracefulStop()
	}()

	select {
	case e := <-errChan:
		logger.GetLoggerFromContext(ctx).Fatal(ctx, "failed to serve", zap.Error(e))
	case <-stopChan:
		logger.GetLoggerFromContext(ctx).Info(ctx, "Shutting down by signal...")
	}
}
