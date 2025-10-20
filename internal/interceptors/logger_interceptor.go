package interceptors

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryLoggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		logger := logger.With(zap.String("method", info.FullMethod))
		ctx = ctxzap.ToContext(ctx, logger)

		logger.Debug("server.request")
		resp, err = handler(ctx, req)
		if err != nil {
			logger.Error("server.request", zap.Error(err))
			return
		}

		return
	}
}

func StreamLoggerInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger := logger.With(zap.String("method", info.FullMethod))
		ctx := ctxzap.ToContext(ss.Context(), logger)
		wrapped := wrapServerStream(ctx, ss)

		logger.Debug("server.stream")
		err := handler(srv, wrapped)
		if err != nil {
			logger.Error("server.stream", zap.Error(err))
			return err
		}

		return nil
	}
}

type ctxDecoratedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *ctxDecoratedServerStream) Context() context.Context {
	return w.ctx
}

func wrapServerStream(ctx context.Context, ss grpc.ServerStream) grpc.ServerStream {
	return &ctxDecoratedServerStream{ServerStream: ss, ctx: ctx}
}
