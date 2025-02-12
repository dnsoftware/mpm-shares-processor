package otel

import (
	"context"
	"log"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Config struct {
	ServiceName        string        // Название сервиса (чтобы понимать с какого сервера идет трассировка и т.п.)
	CollectorEndpoint  string        // адрес:порт Otel коллектора, куда будут отсылаться трассировки
	BatchTimeout       time.Duration // через указанный период времени данные по трассировкам будут отправляться в одном пакете
	MaxExportBatchSize int           // Максимальное кол-во спанов в пакете
	MaxQueueSize       int           // Максимум спанов в очереди
}

// InitTracer Инициализация трассировщика, вызывать в самом начале программы
// Пример вызова:
//
//	cleanup := InitTracer(cfg)
//	defer cleanup()
func InitTracer(cfg Config) func() {

	ctx := context.Background()

	// Создание экспортера OTLP gRPC для отправки данных в OpenTelemetry Collector
	// Создаем OTLP gRPC экспортер для трассировок
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),                      // Без TLS (используйте WithTLS() для включения)
		otlptracegrpc.WithEndpoint(cfg.CollectorEndpoint), // Адрес OpenTelemetry Collector
	)

	if err != nil {
		log.Fatalf("Ошибка создания OTLP gRPC экспортера: %v", err)
	}

	// Создание BatchSpanProcessor
	batchProcessor := trace.NewBatchSpanProcessor(
		exporter,
		trace.WithBatchTimeout(cfg.BatchTimeout), // Отправлять каждые 5 секунд
		trace.WithMaxExportBatchSize(cfg.MaxExportBatchSize), // Не более MaxExportBatchSize спанов за раз
		trace.WithMaxQueueSize(cfg.MaxQueueSize),             // Максимум MaxQueueSize спанов в очереди
	)

	// Оборачиваем BatchSpanProcessor фильтрующим процессором
	filteringProcessor := &filteringSpanProcessor{next: batchProcessor}

	// Настраиваем TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(filteringProcessor),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}
}

// InitSimpleTracer Пример инициализации трассировщика с консольным экспортером:
func InitSimpleTracer() func() {
	// Создаем экспортер для вывода трассировок в консоль
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("failed to initialize stdouttrace exporter: %v", err)
	}

	// Создаем TracerProvider с экспортёром
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter), // Отправка данных в экспортер
		trace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", "example-service"),
		)),
	)

	// Устанавливаем глобальный TracerProvider
	otel.SetTracerProvider(tp)

	// Возвращаем функцию для завершения работы TracerProvider
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}
}

// Для фильтрации спанов
type filteringSpanProcessor struct {
	next trace.SpanProcessor
}

func (fsp *filteringSpanProcessor) OnStart(parent context.Context, span trace.ReadWriteSpan) {
	fsp.next.OnStart(parent, span)
}

// OnEnd Здесь настраиваем фильтры по названию спана
// Пропускаем спаны, связанные с Docker API
func (fsp *filteringSpanProcessor) OnEnd(span trace.ReadOnlySpan) {
	if strings.Contains(span.Name(), "/containers") ||
		strings.Contains(span.Name(), "GET /") ||
		strings.Contains(span.Name(), "HEAD /") {
		return
	}
	fsp.next.OnEnd(span)
}

func (fsp *filteringSpanProcessor) Shutdown(ctx context.Context) error {
	return fsp.next.Shutdown(ctx)
}

func (fsp *filteringSpanProcessor) ForceFlush(ctx context.Context) error {
	return fsp.next.ForceFlush(ctx)
}
