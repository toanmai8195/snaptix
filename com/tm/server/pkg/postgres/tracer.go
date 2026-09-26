package postgres

import (
	"context"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
)

// parentOnlyTracer chỉ tạo span cho truy vấn nằm trong một trace đang có (vd trong request HTTP).
// Truy vấn không có span cha (probe /readyz, job nền chưa gắn trace) bị bỏ qua để không sinh
// hàng loạt trace gốc vô nghĩa.
type parentOnlyTracer struct {
	next pgx.QueryTracer
}

type tracedKey struct{}

func newQueryTracer(opts ...otelpgx.Option) pgx.QueryTracer {
	return parentOnlyTracer{next: otelpgx.NewTracer(opts...)}
}

func (t parentOnlyTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if !trace.SpanContextFromContext(ctx).IsValid() {
		return ctx
	}
	return context.WithValue(t.next.TraceQueryStart(ctx, conn, data), tracedKey{}, true)
}

func (t parentOnlyTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if traced, _ := ctx.Value(tracedKey{}).(bool); traced {
		t.next.TraceQueryEnd(ctx, conn, data)
	}
}
