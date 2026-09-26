package otelx

import "context"

type requestIDKey struct{}

// WithRequestID gắn request ID vào context để log và các lớp dưới đọc được.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFrom đọc request ID từ context; rỗng nếu không có.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
