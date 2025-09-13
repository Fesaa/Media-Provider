package tracing

import (
	"go.opentelemetry.io/otel"
)

const (
	tracerPrefix = "github.com/Fesaa/Media-Provider/"
)

var (
	TracerMain     = otel.Tracer(tracerPrefix + "main")
	TracerApi      = otel.Tracer(tracerPrefix + "api")
	TracerDb       = otel.Tracer(tracerPrefix + "db")
	TracerPasloe   = otel.Tracer(tracerPrefix + "providers/pasloe")
	TracerServices = otel.Tracer(tracerPrefix + "services")
)

const (
	SpanPasloeCachedDownload  = "pasloe.download.cached"
	SpanPasloeDownloadContent = "pasloe.download.content"
	SpanPasloeIOWorker        = "pasloe.download.io_worker"
	SpanPasloeDownloadWorker  = "pasloe.download.worker"
	SpanPasloeChapter         = "pasloe.download.chapter"
	SpanPasloeRatelimit       = "pasloe.download.ratelimit"
	SpanPasloeLoadMetadata    = "pasloe.metadata.load"
	SpanPasLoadContentInfo    = "pasloe.metadata.load.info"
	SpanPasloeContentFilter   = "pasloe.metadata.filter.content"
	SpanPasloeCovers          = "pasloe.metadata.covers"

	SpanServicesImagesWebp       = "services.images.covert.webp"
	SpanServicesTranslocoLoading = "services.transloco.loading"
	SpanServicesCache            = "services.cache"
	SpanServicesContentSearch    = "services.content.search"
	SpanServicesSubscriptionTask = "services.subscription.task"
	SpanServicesOIDCTokenRefresh = "services.auth.oidc.token_refresh" //nolint: gosec

	SpanApplicationStart = "application.start"
	SpanUpdateVersion    = "application.start.version_update"
	SpanRegisterApi      = "application.start.register_api"
	SpanMigrations       = "application.start.migrations"
	SpanManualMigrations = "application.start.manual_migrations"
	SpanSetupDb          = "application.start.setup_db"
	SpanSetupService     = "application.start.setup_service"
	SetupRepository      = "application.start.setup_repository"
)
