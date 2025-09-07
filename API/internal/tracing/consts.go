package tracing

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerPasloe   = "github.com/Fesaa/Media-Provider/providers/pasloe"
	TracerServices = "github.com/Fesaa/Media-Provider/services"
)

func PasloeTracer() trace.Tracer {
	return otel.Tracer(TracerPasloe)
}

func ServicesTracer() trace.Tracer {
	return otel.Tracer(TracerServices)
}

const (
	SpanPasloeDownloadContent = "pasloe.download.content"
	SpanPasloeIOWorker        = "pasloe.download.io_worker"
	SpanPasloeDownloadWorker  = "pasloe.download.worker"
	SpanPasloeChapter         = "pasloe.download.chapter"

	SpanServicesImagesWebp = "services.images.covert.webp"
)
