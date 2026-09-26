# Test cases — P0-T10: OpenTelemetry SDK trong pkg/otelx, export OTLP

> Viết ở bước 1. **Tự duyệt** theo chỉ đạo người dùng (Phase 0 chưa có logic nghiệp vụ). Task: xem [README phase](../../README.md).

`otelx.Setup(ctx, Config)` khởi tạo TracerProvider + MeterProvider (OTLP HTTP), propagator W3C, trả hàm `shutdown` để flush khi dừng. PG query được trace bằng otelpgx, chỉ khi có span cha.

| ID | Loại | Kịch bản | Kết quả mong đợi | Trạng thái |
|---|---|---|---|---|
| P0-T10-TC01 | Unit | `Setup` với endpoint là server OTLP giả (httptest), tạo span + ghi metric, gọi `shutdown` | Server giả nhận `POST /v1/traces` và `POST /v1/metrics`; resource có `service.name=core`, `service.version` | ✅ |
| P0-T10-TC02 | Unit | `Config.Disabled` (`OTEL_SDK_DISABLED=true`) | Không gửi request nào; tracer/meter no-op; `shutdown` trả `nil` | ✅ |
| P0-T10-TC03 | Unit | Sau `Setup`, propagator global | Inject/extract được `traceparent` và `baggage` | ✅ |
| P0-T10-TC04 | Unit | Collector không truy cập được | `Setup` không lỗi và không chặn; `shutdown` trả về trong thời hạn ctx (≤ 3s), không panic | ✅ |
| P0-T10-TC05 | Unit | Tracer PG chỉ khi có span cha | Truy vấn có span cha → có span con (DB); truy vấn không có span cha (vd probe `/readyz`) → không tạo span | ✅ |
| P0-T10-TC06 | Script | Stack (collector, Tempo, Prometheus) + binary core, gửi request có `traceparent` | Tempo có trace đó với span của service `core`; Prometheus có `http_server_request_duration_seconds_count{service_name="core"}` | ✅ |

Test nghiệm thu liên quan: P0-AT07 (trace bff → core → PG — cần P0-T12); challenge **G10**.
