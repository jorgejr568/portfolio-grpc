package interceptors

import (
	"context"
	"strings"

	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
	"google.golang.org/grpc"
)

// StatsDInterceptor creates a gRPC unary interceptor that tracks metrics for each request
func StatsDInterceptor(statsdClient statsd.Client) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Parse service and method from full method name
		// Format: /package.ServiceName/MethodName
		serviceName, methodName := parseMethodName(info.FullMethod)

		// Start tracking the request
		tracker := statsdClient.Start(serviceName, methodName)

		// Call the actual handler
		resp, err := handler(ctx, req)

		// Track the result
		if err != nil {
			_ = tracker.FailedWithError(err)
		} else {
			_ = tracker.Succeeded()
		}

		return resp, err
	}
}

// parseMethodName extracts service and method names from gRPC full method
// Example: "/jorgejr568.portfolio_grpc.PortfolioService/GetAllSkills" -> ("PortfolioService", "GetAllSkills")
func parseMethodName(fullMethod string) (serviceName, methodName string) {
	// Remove leading slash
	fullMethod = strings.TrimPrefix(fullMethod, "/")

	// Split by slash to separate package.service from method
	parts := strings.Split(fullMethod, "/")
	if len(parts) != 2 {
		return "unknown", "unknown"
	}

	// Extract method name
	methodName = parts[1]

	// Extract service name from package.service
	serviceParts := strings.Split(parts[0], ".")
	if len(serviceParts) > 0 {
		serviceName = serviceParts[len(serviceParts)-1]
	} else {
		serviceName = "unknown"
	}

	return serviceName, methodName
}
