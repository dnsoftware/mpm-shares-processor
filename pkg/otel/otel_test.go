package otel

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Тест трассировки в памяти
func TestTracing(t *testing.T) {
	// Создаем in-memory экспортер
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter), // Записываем спаны в память
	)
	defer tp.Shutdown(context.Background())

	otel.SetTracerProvider(tp)

	// Трассировка кода
	tracer := otel.Tracer("test-tracer")
	ctx, span := tracer.Start(context.Background(), "TestOperation")
	_ = ctx
	span.SetAttributes(attribute.String("test.key", "test_value"))
	span.End()

	// Проверяем данные из экспортёра
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	// Проверяем содержимое спана
	spanData := spans[0]
	if spanData.Name != "TestOperation" {
		t.Errorf("expected span name 'TestOperation', got '%s'", spanData.Name)
	}

	// Проверяем содержимое атрибутов спана
	found := false
	for _, attr := range spanData.Attributes {
		if attr.Key == "test.key" && attr.Value.AsString() == "test_value" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected attribute 'test.key' with value 'test_value', but not found")
	}

}
