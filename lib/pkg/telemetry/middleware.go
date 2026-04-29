package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"


	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)


var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"method", "status"},
	)
)

// GinMiddleware returns a Gin middleware that records Prometheus metrics
// and structured request logs with trace ID correlation.
func GinMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record Metrics
		requestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)

		// Extract Trace ID
		span := trace.SpanFromContext(c.Request.Context())
		traceID := ""
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
		}

		// Structured Logging
		logger.Infow("request completed",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration_seconds", duration,
			"trace_id", traceID,
			"client_ip", c.ClientIP(),
		)
	}
}

// GrpcInterceptor returns a gRPC unary interceptor for logging and metrics.
func GrpcInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		status := "ok"
		if err != nil {
			status = "error"
		}

		// Record Metrics
		grpcRequestsTotal.WithLabelValues(info.FullMethod, status).Inc()

		// Extract Trace ID
		span := trace.SpanFromContext(ctx)
		traceID := ""
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
		}

		logger.Infow("grpc request completed",
			"method", info.FullMethod,
			"status", status,
			"duration_seconds", duration,
			"trace_id", traceID,
		)

		return resp, err
	}
}

// ExposeMetrics starts a metrics server on the given port.
func ExposeMetrics(port string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	return server.ListenAndServe()
}


