package grpc

import (
	"context"
	"github.com/anatoly_dev/go-users/app/pkg/metrics"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsInterceptor struct {
	metricsHelper *metrics.MetricsHelper
}

func NewMetricsInterceptor(metricsHelper *metrics.MetricsHelper) *MetricsInterceptor {
	return &MetricsInterceptor{
		metricsHelper: metricsHelper,
	}
}

func (m *MetricsInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		m.metricsHelper.GRPCConcurrentRequests().Inc()
		defer m.metricsHelper.GRPCConcurrentRequests().Dec()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		method := m.extractMethodName(info.FullMethod)
		statusCode := m.getStatusCode(err)

		requestSize := m.estimateSize(req)
		responseSize := m.estimateSize(resp)

		m.metricsHelper.RecordGRPCRequest(
			method,
			statusCode,
			duration,
			requestSize,
			responseSize,
		)

		if err != nil {
			errorType := m.categorizeGRPCError(err)
			m.metricsHelper.ErrorsTotal().WithLabelValues("grpc", errorType).Inc()
		}

		if duration.Seconds() > 1.0 {
			m.metricsHelper.SlowResponseTime().WithLabelValues("1s").Set(1)
		}

		m.metricsHelper.ResponseTimePercentiles().WithLabelValues("grpc").Observe(duration.Seconds())

		return resp, err
	}
}

func (m *MetricsInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		m.metricsHelper.GRPCConcurrentRequests().Inc()
		defer m.metricsHelper.GRPCConcurrentRequests().Dec()

		wrappedStream := &metricsServerStream{
			ServerStream:  ss,
			metricsHelper: m.metricsHelper,
		}

		err := handler(srv, wrappedStream)

		duration := time.Since(start)
		method := m.extractMethodName(info.FullMethod)
		statusCode := m.getStatusCode(err)

		m.metricsHelper.RecordGRPCRequest(
			method,
			statusCode,
			duration,
			0,
			0,
		)

		if err != nil {
			errorType := m.categorizeGRPCError(err)
			m.metricsHelper.ErrorsTotal().WithLabelValues("grpc", errorType).Inc()
		}

		return err
	}
}

func (m *MetricsInterceptor) PanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				m.metricsHelper.RecordPanicRecovery()
				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

type metricsServerStream struct {
	grpc.ServerStream
	metricsHelper *metrics.MetricsHelper
}

func (m *metricsServerStream) SendMsg(msg interface{}) error {
	err := m.ServerStream.SendMsg(msg)
	if err == nil {
		m.metricsHelper.GRPCStreamMessagesSent().Inc()
	}
	return err
}

func (m *metricsServerStream) RecvMsg(msg interface{}) error {
	err := m.ServerStream.RecvMsg(msg)
	if err == nil {
		m.metricsHelper.GRPCStreamMessagesReceived().Inc()
	}
	return err
}

func (m *MetricsInterceptor) extractMethodName(fullMethod string) string {
	if len(fullMethod) == 0 {
		return "unknown"
	}

	lastSlash := -1
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || lastSlash == len(fullMethod)-1 {
		return "unknown"
	}

	return fullMethod[lastSlash+1:]
}

func (m *MetricsInterceptor) getStatusCode(err error) string {
	if err == nil {
		return "OK"
	}

	st, ok := status.FromError(err)
	if !ok {
		return "UNKNOWN"
	}

	return st.Code().String()
}

func (m *MetricsInterceptor) categorizeGRPCError(err error) string {
	if err == nil {
		return "none"
	}

	st, ok := status.FromError(err)
	if !ok {
		return "unknown"
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return "invalid_argument"
	case codes.NotFound:
		return "not_found"
	case codes.AlreadyExists:
		return "already_exists"
	case codes.PermissionDenied:
		return "permission_denied"
	case codes.Unauthenticated:
		return "unauthenticated"
	case codes.ResourceExhausted:
		return "resource_exhausted"
	case codes.FailedPrecondition:
		return "failed_precondition"
	case codes.Aborted:
		return "aborted"
	case codes.OutOfRange:
		return "out_of_range"
	case codes.Unimplemented:
		return "unimplemented"
	case codes.Internal:
		return "internal"
	case codes.Unavailable:
		return "unavailable"
	case codes.DataLoss:
		return "data_loss"
	case codes.DeadlineExceeded:
		return "deadline_exceeded"
	case codes.Canceled:
		return "cancelled"
	default:
		return "unknown"
	}
}

func (m *MetricsInterceptor) estimateSize(msg interface{}) int64 {
	if msg == nil {
		return 0
	}

	return 100
}
