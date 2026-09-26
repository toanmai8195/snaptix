package postgres

import (
	"context"
	"testing"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TC05
func TestParentOnlyTracer(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	tracer := newQueryTracer(otelpgx.WithTracerProvider(tp))

	query := func(ctx context.Context) {
		ctx = tracer.TraceQueryStart(ctx, nil, pgx.TraceQueryStartData{SQL: "SELECT 1"})
		tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{})
	}

	// Không có span cha (vd probe /readyz): không tạo span.
	query(context.Background())
	if n := len(rec.Ended()); n != 0 {
		t.Fatalf("truy vấn không có span cha tạo %d span, muốn 0", n)
	}

	// Có span cha (trong request HTTP): tạo span con cùng trace.
	ctx, parent := tp.Tracer("test").Start(context.Background(), "GET /x")
	query(ctx)
	parent.End()

	ended := rec.Ended()
	if len(ended) != 2 {
		t.Fatalf("muốn 2 span (query + cha), có %d", len(ended))
	}
	child := ended[0]
	if child.Parent().SpanID() != parent.SpanContext().SpanID() || child.SpanContext().TraceID() != parent.SpanContext().TraceID() {
		t.Errorf("span query phải là con của span request")
	}
}
