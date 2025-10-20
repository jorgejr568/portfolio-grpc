package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
	"github.com/jorgejr568/portfolio-grpc/internal/env"
	"github.com/jorgejr568/portfolio-grpc/internal/interceptors"
	"github.com/jorgejr568/portfolio-grpc/internal/repositories"
	"github.com/jorgejr568/portfolio-grpc/internal/server"
	_ "github.com/lib/pq"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	statsdPrefix   = "portfolio_grpc"
	statsdEnv      = "STATSD_ADDRESS"
	databaseURLEnv = "DATABASE_URL"
	grpcPort       = ":50051"
	httpPort       = ":8080"
)

func main() {
	ctx := context.Background()
	_ = env.LoadEnv()
	if err := env.ValidateEnv(); err != nil {
		log.Fatalf("environment validation failed: %v", err)
	}

	di := dig.New()
	err := di.Provide(func() (*zap.Logger, error) {
		logLevelStr := os.Getenv("LOG_LEVEL")
		if logLevelStr == "" {
			logLevelStr = "info"
		}

		var logLevel zapcore.Level
		if err := logLevel.UnmarshalText([]byte(logLevelStr)); err != nil {
			return nil, fmt.Errorf("invalid LOG_LEVEL: %w", err)
		}

		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(logLevel)

		logger, err := cfg.Build()
		if err != nil {
			return nil, err
		}

		return logger, nil
	})

	defer func() {
		_ = di.Invoke(func(logger *zap.Logger) {
			_ = logger.Sync()
		})
	}()

	err = di.Provide(func() (*sql.DB, error) {
		db, err := sql.Open("postgres", os.Getenv(databaseURLEnv))
		if err != nil {
			return nil, err
		}

		err = db.Ping()
		if err != nil {
			return nil, err
		}

		return db, nil
	})
	if err != nil {
		log.Fatalf("failed to provide database to DI container: %v", err)
	}

	err = di.Provide(func() (statsd.Client, error) {
		statsdAddress := os.Getenv(statsdEnv)
		if statsdAddress == "" {
			return statsd.Nop(), nil
		}

		parts := strings.Split(statsdAddress, ":")
		if len(parts) != 2 {
			return nil, errors.New("invalid statsd_address format")
		}

		host := parts[0]
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.New("invalid statsd_address port")
		}

		return statsd.New(statsd.Config{
			Host:   host,
			Port:   port,
			Prefix: statsdPrefix,
		})
	})

	err = registerRepositories(di)
	if err != nil {
		log.Fatalf("failed to register repositories in DI container: %v", err)
	}

	err = di.Provide(server.NewServer)
	if err != nil {
		log.Fatalf("failed to provide Server to DI container: %v", err)
	}

	err = di.Invoke(func(srv server.Server, logger *zap.Logger, st statsd.Client) error {
		// Create context that listens for interrupt signals
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Setup signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		// Start gRPC server
		grpcServer := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				interceptors.UnaryLoggerInterceptor(logger),
				interceptors.StatsDInterceptor(st),
			),
			grpc.ChainStreamInterceptor(
				interceptors.StreamLoggerInterceptor(logger),
			),
		)

		go func() {
			if err := startGRPCServer(ctx, grpcServer, srv, logger); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				logger.Fatal("gRPC server error: %v", zap.Error(err))
			}
		}()

		// Start HTTP Gateway server
		httpServer := &http.Server{Addr: httpPort}
		go func() {
			if err := startHTTPGateway(ctx, httpServer, srv, logger); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("HTTP gateway server error: %v", zap.Error(err))
			}
		}()

		// Wait for interrupt signal
		<-sigChan
		logger.Info("⚠️  Shutdown signal received, stopping servers...")

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// Shutdown HTTP server
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error: %v", zap.Error(err))
		}

		// Gracefully stop gRPC server
		grpcServer.GracefulStop()

		logger.Info("✅ Servers stopped successfully")
		return nil
	})

	if err != nil {
		_ = di.Invoke(func(logger *zap.Logger) error {
			logger.Fatal("failed to start servers", zap.Error(err))
			return nil
		})
	}
}

func startGRPCServer(ctx context.Context, grpcServer *grpc.Server, srv server.Server, logger *zap.Logger) error {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	portfolio_grpc.RegisterPortfolioServiceServer(grpcServer, srv)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	reflection.Register(grpcServer)

	logger.Info(fmt.Sprintf("✅ gRPC server listening on %s", grpcPort))

	errChan := make(chan error, 1)
	// Serve in a goroutine so we can handle shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		if err := lis.Close(); err != nil {
			return err
		}

		return nil
	case err := <-errChan:
		return err
	}
}

func startHTTPGateway(ctx context.Context, httpServer *http.Server, srv server.Server, logger *zap.Logger) error {
	// Create gRPC-Gateway mux
	mux := runtime.NewServeMux()

	conn, err := grpc.NewClient(
		"127.0.0.1"+grpcPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	defer conn.Close()
	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if ready := conn.WaitForStateChange(connCtx, connectivity.Ready); !ready {
		return fmt.Errorf("failed to establish connection to gRPC server")
	}

	// Register using the pre-established connection
	err = portfolio_grpc.RegisterPortfolioServiceHandlerClient(
		ctx,
		mux,
		portfolio_grpc.NewPortfolioServiceClient(conn),
	)
	if err != nil {
		return err
	}

	// Add CORS middleware
	httpServer.Handler = corsMiddleware(mux)

	logger.Info(fmt.Sprintf("✅ HTTP/REST gateway listening on %s", httpPort))
	logger.Info("REST API Endpoints:", zap.Strings("endpoints", []string{
		"GET  http://localhost:8080/v1/skills",
		"GET  http://localhost:8080/v1/skills/{id}",
		"GET  http://localhost:8080/v1/experiences",
		"GET  http://localhost:8080/v1/experiences/{id}",
		"GET  http://localhost:8080/v1/educations",
		"GET  http://localhost:8080/v1/educations/{id}",
	}))

	return httpServer.ListenAndServe()
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func registerRepositories(di *dig.Container) error {
	if err := di.Provide(repositories.NewSkillsRepository); err != nil {
		return err
	}

	if err := di.Provide(repositories.NewExperiencesRepository); err != nil {
		return err
	}

	if err := di.Provide(repositories.NewEducationsRepository); err != nil {
		return err
	}

	return nil
}
