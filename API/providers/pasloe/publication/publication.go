package publication

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

const (
	DownloadOneShotKey       string = "download_one_shot"
	IncludeNotMatchedTagsKey string = "include_not_matched_tags"
	IncludeCover             string = "include_cover"
	UpdateCover              string = "update_cover"
	TitleOverride            string = "title_override"
	AssignEmptyVolumes       string = "assign_empty_volumes"
	ScanlationGroupKey       string = "scanlation_group"
	SkipVolumeWithoutChapter string = "skip_volume_without_chapter"
)

const (
	toggleSubscriptionExhausted = "toggle_subscription_exhausted"
	togglePreferencesFailed     = "toggle_blacklist_failed"
)

type Client interface {
	services.Client
	GetBaseDir() string
	MoveToDownloadQueue(id string) error
	GetCurrentDownloads() []Publication
}

type Publication interface {
	services.Content

	Cancel()
	GetDownloadDir() string

	DisplayInformation() string

	// GetOnDiskContent returns the name of the files that have been identified as already existing content
	GetOnDiskContent() []Content
	// GetNewContent returns the full (relative) path of downloaded content.
	// This will be a slice of paths produced by DownloadInfoProvider.ContentPath
	GetNewContent() []string
	// GetToRemoveContent returns the full (relative) path of old content that has to be removed
	GetToRemoveContent() []string
	// CleanupNewContent takes a path from GetNewContent to clean up
	CleanupNewContent(string) error

	// LoadMetadata loads all required metadata to start the download , this method blocks until complete or cancelled
	LoadMetadata(ctx context.Context)
	// DownloadContent starts the download process, this method blocks until complete or cancelled
	DownloadContent(ctx context.Context)

	// GetNewContentNamed returns the names of the downloaded content (chapters)
	GetNewContentNamed() []string

	FailedDownloads() int
	UpdateSeriesInfo(f func(*Series))
}

type Content struct {
	Name string
	Path string
	// The following fields are parsed from the file name
	Chapter string
	Volume  string
}

func New(
	log zerolog.Logger,
	signalR services.SignalRService,
	notificationService services.NotificationService,
	translocoService services.TranslocoService,
	archiveService services.ArchiveService,
	imageService services.ImageService,
	dirService services.DirectoryService,
	settingsService services.SettingsService,
	unitOfWork *db.UnitOfWork,
	fs afero.Afero,
	httpClient *menou.Client,
	client Client,
	req payload.DownloadRequest,
	repository Repository,
	ext Extensions,
) (Publication, error) {

	settings, err := settingsService.GetSettingsDto(context.Background())
	if err != nil {
		return nil, err
	}

	return &publication{
		log:                 log.With().Str("handler", req.Provider.String()).Logger(),
		signalR:             signalR,
		notificationService: notificationService,
		translocoService:    translocoService,
		archiveService:      archiveService,
		imageService:        imageService,
		dirService:          dirService,
		unitOfWork:          unitOfWork,
		fs:                  fs,
		httpClient:          httpClient,
		client:              client,
		ext:                 ext,

		maxImages:  utils.Clamp(settings.MaxConcurrentImages, 1, 5),
		req:        req,
		repository: repository,

		toggles:               utils.NewToggles[string](),
		hasDuplicatedChapters: utils.Settable[bool]{},
		speedTracker:          utils.NewSpeedTracker(0),
	}, nil
}

type publication struct {
	log                 zerolog.Logger
	signalR             services.SignalRService
	notificationService services.NotificationService
	translocoService    services.TranslocoService
	archiveService      services.ArchiveService
	imageService        services.ImageService
	dirService          services.DirectoryService
	unitOfWork          *db.UnitOfWork
	fs                  afero.Afero
	httpClient          *menou.Client
	client              Client
	ext                 Extensions

	maxImages   int
	req         payload.DownloadRequest
	state       payload.ContentState
	preferences *models.UserPreferences
	repository  Repository
	series      *Series

	toggles *utils.Toggles[string]
	// hasDuplicatedChapters is true if the same chapter number is used across different volumes
	// forcing us to use volumes in the file name
	hasDuplicatedChapters utils.Settable[bool]

	toDownload      []string  // All chapters id that need to be downloaded
	hasDownloaded   []string  // path of files on disk we've already downloaded
	existingContent []Content // Content already on disk before download started
	toRemoveContent []string  // content on disk that has to be removed as it has been redownloaded

	// toDownloadUserSelected are the ids of the content selected by the user to download in the UI
	toDownloadUserSelected []string

	failedDownloads int64
	speedTracker    *utils.SpeedTracker

	cancel context.CancelFunc
	// Wait group used to track chapters being downloaded
	wg *sync.WaitGroup

	// Wait group for IO workers
	ioWg     *sync.WaitGroup
	iOWorkCh chan ioTask
}

func (p *publication) Id() string {
	if p.series != nil {
		return p.series.Id
	}

	return p.req.Id
}

func (p *publication) Title() string {
	if p.series != nil {
		return utils.NonEmpty(p.req.GetStringOrDefault(TitleOverride, ""), p.series.Title)
	}

	return utils.NonEmpty(p.req.GetStringOrDefault(TitleOverride, ""), p.req.TempTitle, p.req.Id)
}

func (p *publication) Provider() models.Provider {
	return p.req.Provider
}

func (p *publication) State() payload.ContentState {
	return p.state
}

func (p *publication) SetState(state payload.ContentState) {
	p.state = state
	p.signalR.StateUpdate(p.req.OwnerId, p.Id(), p.state)
}

func (p *publication) Request() payload.DownloadRequest {
	return p.req
}

func (p *publication) Cancel() {
	p.log.Trace().Msg("download of the publication is being cancelled")

	if p.cancel != nil {
		p.cancel()
	}

	if p.wg != nil {
		p.log.Debug().Msg("waiting at most 1 minute for all download task to complete")
		utils.WaitFor(p.wg, time.Minute)
		p.log.Debug().Msg("download tasks completed, waiting at most 1 minute for I/O")
		utils.WaitFor(p.ioWg, time.Minute)
	}

	if p.client.Content(p.Id()) == nil {
		return
	}

	req := payload.StopRequest{
		Provider:    p.Provider(),
		Id:          p.Id(),
		DeleteFiles: true,
	}
	if err := p.client.RemoveDownload(req); err != nil {
		p.log.Warn().Err(err).Msg("failed to remove download")
	}
}

func (p *publication) GetDownloadDir() string {
	if p.series != nil {
		return path.Join(p.req.BaseDir, p.Title())
	}

	return p.req.BaseDir
}

func (p *publication) DisplayInformation() string {
	if p.req.IsSubscription && p.req.Sub != nil {
		return p.req.Sub.Title
	}

	return p.Title()
}

func (p *publication) GetOnDiskContent() []Content {
	return p.existingContent
}

func (p *publication) GetNewContent() []string {
	return p.hasDownloaded
}

func (p *publication) GetContentByName(name string) (Content, bool) {
	for _, content := range p.existingContent {
		if strings.TrimSuffix(content.Name, path.Ext(content.Name)) == name { // ignore file ext
			return content, true
		}
	}
	return Content{}, false
}

func (p *publication) GetContentByPath(path string) (Content, bool) {
	for _, content := range p.existingContent {
		if content.Path == path {
			return content, true
		}
	}
	return Content{}, false
}

func (p *publication) GetContentByVolumeAndChapter(volume string, chapter string) (Content, bool) {
	for _, content := range p.existingContent {
		if content.Volume == volume && content.Chapter == chapter {
			return content, true
		}

		// Content has been assigned a volume
		if content.Volume == "" && content.Chapter == chapter {
			return content, true
		}

		// Content has had its volume removed
		if content.Volume != "" && volume == "" && content.Chapter == chapter {
			return content, true
		}
	}
	return Content{}, false
}

func (p *publication) GetToRemoveContent() []string {
	return p.toRemoveContent
}

func (p *publication) CleanupNewContent(s string) error {
	return p.ext.contentCleanupFunc(p, s)
}

// StopDownload sends a payload.StopRequest to the Client with deletefiles=false
func (p *publication) StopDownload() {
	if err := p.client.RemoveDownload(payload.StopRequest{Provider: p.req.Provider, Id: p.Id()}); err != nil {
		p.log.Error().Err(err).Msg("error while cleaning up")
	}
}

func (p *publication) LoadMetadata(ctx context.Context) {
	if p.cancel != nil {
		p.log.Warn().Msg("content is already loading info, or downloading")
		return
	}

	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeLoadMetadata)
	defer span.End()

	p.log.Debug().Msg("loading content info")
	p.SetState(payload.ContentStateLoading)

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	start := time.Now()

	up, err := p.unitOfWork.Preferences.GetPreferences(ctx, p.req.OwnerId)
	if err != nil {
		p.log.Error().Err(err).Msg("unable to get preferences, some features may not work")
	}
	p.preferences = up

	if err = p.loadSeriesInfo(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			p.log.Error().Err(err).Msg("failed to load series info")
		}
		p.StopDownload()
		return
	}

	// We use the nil check rather than isSub as subs triggered by the one time flow. You go over this logic
	if p.req.Sub != nil {

		if p.req.Sub.Payload.Extra == nil {
			p.req.Sub.Payload.Extra = map[string][]string{}
		}

		if p.req.Sub.Payload.Extra.GetStringOrDefault(TitleOverride, "") == "" {
			p.req.Sub.Payload.Extra.SetValue(TitleOverride, p.Title())
			if err = p.unitOfWork.Subscriptions.Update(ctx, *p.req.Sub); err != nil {
				p.log.Warn().Err(err).Msg("failed to set title override")
			}
		}

		if err = p.ensureSubscriptionDirectoryIsUpToDate(ctx); err != nil {
			p.log.Error().Err(err).Msg("An error occurred while updating subscription directories. Cancelling download")
			p.StopDownload()
			return
		}
	}

	ctx, span = tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeContentFilter)
	defer span.End()

	dur, err := p.filterAlreadyDownloadedContent(ctx)
	if err != nil {
		p.Cancel()
		return
	}

	if dur > time.Second*5 {
		p.log.Warn().Dur("elapsed", dur).Msg("checking for on disk content took a long time")

		if p.req.IsSubscription {
			p.notificationService.Notify(ctx, models.NewNotification().
				WithTitle(p.translocoService.GetTranslation("warn")).
				WithSummary(p.translocoService.GetTranslation("long-on-disk-check", p.Title())).
				WithBody(p.translocoService.GetTranslation("long-on-disk-check-body", dur)).
				WithGroup(models.GroupContent).
				WithColour(models.Warning).
				WithOwner(p.req.OwnerId).
				WithRequiredRoles(models.ViewAllDownloads).
				Build())
		}
	}

	p.handleSubscriptionNoDownloadCount(ctx, len(p.toDownload) > 0)

	if len(p.toDownload) == 0 && p.req.DownloadMetadata.StartImmediately {
		p.log.Debug().Msg("no chapters found to download, stopping")
		p.SetState(payload.ContentStateWaiting)
		p.StopDownload()
		return
	}

	if p.req.DownloadMetadata.StartImmediately {
		p.SetState(payload.ContentStateReady)
	} else {
		p.SetState(payload.ContentStateWaiting)
	}
	p.signalR.UpdateContentInfo(p.req.OwnerId, p.GetInfo())

	p.log.Debug().
		Dur("duration", time.Since(start)).
		Int("allChapters", len(p.series.Chapters)).
		Int("toDownload", len(p.toDownload)).
		Msg("loaded content info")
}

func (p *publication) handleSubscriptionNoDownloadCount(ctx context.Context, reset bool) {
	if !p.req.IsSubscription {
		return
	}

	if reset {
		p.req.Sub.NoDownloadCount = 0
	} else {
		p.req.Sub.NoDownloadCount++
	}

	// Leave notifications for counts above 5, and only once every 5 days
	if p.req.Sub.NoDownloadCount >= 5 && p.req.Sub.NoDownloadCount%5 == 0 && p.preferences.LogEmptyDownloads {
		p.notificationService.Notify(ctx, models.NewNotification().
			WithTitle(p.translocoService.GetTranslation("sub-too-frequent")).
			WithBody(p.translocoService.GetTranslation("sub-too-frequent-body", p.Title())).
			WithGroup(models.GroupContent).
			WithColour(models.Warning).
			WithOwner(p.req.OwnerId).
			WithRequiredRoles(models.ViewAllDownloads).
			Build())
	}

	if err := p.unitOfWork.Subscriptions.Update(ctx, *p.req.Sub); err != nil {
		p.log.Warn().Err(err).Msg("failed to update no download count for subscription")
	}
}

func (p *publication) loadSeriesInfo(ctx context.Context) error {
	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasLoadContentInfo)
	defer span.End()

	start := time.Now()

	series, err := p.repository.SeriesInfo(ctx, p.Id(), p.req)
	if err != nil {
		return err
	}

	p.log.Debug().Dur("duration", time.Since(start)).Msg("loaded series info")

	if series.Title == "" {
		return fmt.Errorf("no title found in series info")
	}

	p.series = &series
	p.log = p.log.With().Str("title", p.Title()).Logger()

	if group, ok := p.req.GetString(ScanlationGroupKey); ok {
		p.log.Debug().Str("group", group).Msg("filtering chapters on translator")
		p.series.Chapters = utils.Filter(p.series.Chapters, func(chapter Chapter) bool {
			return slices.Contains(chapter.Translator, group)
		})
	}

	if !p.req.GetBool(AssignEmptyVolumes, false) {
		return nil
	}

	hasVolumes := utils.Any(p.series.Chapters, func(chapter Chapter) bool {
		return chapter.Volume != ""
	})
	if !hasVolumes {
		return nil
	}

	p.series.Chapters = utils.Map(p.series.Chapters, func(chapter Chapter) Chapter {
		if chapter.Volume == "" && chapter.Chapter != "" {
			chapter.Volume = "1"
		}
		return chapter
	})

	return nil
}

// ensureSubscriptionDirectoryIsUpToDate moves previous directories a subscription was downloaded it, in case it changed
// due to upstream changes
func (p *publication) ensureSubscriptionDirectoryIsUpToDate(ctx context.Context) error {
	if p.req.Sub == nil {
		return nil
	}

	var (
		oldDir = p.req.Sub.LastDownloadDir
		newDir = path.Join(p.client.GetBaseDir(), p.GetDownloadDir())
	)

	if p.req.Sub.LastDownloadDir == "" {
		p.log.Debug().Msg("No previous download directory known, updating to current")
		p.req.Sub.LastDownloadDir = newDir
		return p.unitOfWork.Subscriptions.Update(ctx, *p.req.Sub)
	}

	if oldDir == newDir {
		return nil
	}

	p.log.Warn().
		Str("prev-dir", oldDir).
		Str("current-dir", newDir).
		Msg("Download directory out of sync for subscription, moving content")

	ok, err := p.fs.DirExists(newDir)
	if err != nil {
		return err
	}

	if ok {
		p.log.Error().
			Str("dir", newDir).
			Msg("New directory already exists, cannot move old content, aborting download")
		return fs.ErrExist
	}

	return p.fs.Rename(oldDir, newDir)
}

// Only context.Canceled is returned as error
func (p *publication) filterAlreadyDownloadedContent(ctx context.Context) (time.Duration, error) {
	start := time.Now()

	var err error
	p.existingContent, err = p.onDiskContent(ctx)
	if err != nil {
		return time.Since(start), err
	}

	p.toDownload = utils.MaybeMap(p.series.Chapters, func(chapter Chapter) (string, bool) {
		return chapter.Id, p.ShouldDownload(chapter)
	})
	return time.Since(start), nil
}

// onDiskContent returns all content found on disk if no error is found
// When given context.Context is cancelled, at most one more file will be checked
// Only context.Canceled is returned as error
func (p *publication) onDiskContent(ctx context.Context) ([]Content, error) {
	p.log.Debug().Str("dir", p.GetDownloadDir()).Msg("checking content on disk")
	content, err := p.parseDirectoryForContent(ctx, p.GetDownloadDir())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}

		if errors.Is(err, os.ErrNotExist) {
			p.log.Trace().Msg("download directory not found, fresh download")
			return []Content{}, nil
		}

		p.log.Warn().Err(err).Msg("failed to load already downloaded content, all content will be downloaded")
		return []Content{}, nil
	}

	p.log.Trace().Str("content", fmt.Sprintf("%v", content)).Msg("found content on disk")
	return content, nil
}

// parseDirectoryForContent parses the directory @ Client.GetBaseDir/contentPath for conent
// A file is considered content if PublicationExtensions.isContentFunc return true
// The given context.Context is checked for an error on each loop per entry
func (p *publication) parseDirectoryForContent(ctx context.Context, contentPath string) ([]Content, error) {
	dirEntries, err := p.fs.ReadDir(path.Join(p.client.GetBaseDir(), contentPath))
	if err != nil {
		return nil, err
	}

	content := make([]Content, 0)
	for _, entry := range dirEntries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if entry.IsDir() {
			dirContent, err := p.parseDirectoryForContent(ctx, path.Join(contentPath, entry.Name()))
			if err != nil {
				return nil, err
			}
			content = append(content, dirContent...)
			continue
		}

		c, ok := p.ext.isContentFunc(entry.Name())
		if !ok {
			p.log.Trace().Str("fileName", entry.Name()).Msg("skipping non-content file")
			continue
		}

		p.log.Trace().Str("fileName", entry.Name()).Msg("adding content to list")
		content = append(content, Content{
			Name:    entry.Name(),
			Path:    path.Join(contentPath, entry.Name()),
			Chapter: c.Chapter,
			Volume:  c.Volume,
		})
	}

	return content, nil
}

// ShouldDownload returns true if the given chapter should be downloaded, this may be overwritten by a user selection
func (p *publication) ShouldDownload(chapter Chapter) bool {
	// Backwards compatibility check if volume has been downloaded
	if _, ok := p.GetContentByName(p.VolumeDir(chapter)); ok {
		return false
	}

	content, ok := p.GetContentByName(p.ContentFileName(chapter))
	if !ok {
		content, ok = p.GetContentByVolumeAndChapter(chapter.Volume, chapter.Chapter)
		if !ok {
			// Some providers, *dynasty*, have terrible naming schemes for specials.
			if p.req.GetBool(SkipVolumeWithoutChapter, false) && chapter.Volume != "" {
				return chapter.Chapter != ""
			}

			return true
		}
	}

	onDiskVolume, err := p.ext.volumeFunc(p, content)
	if err != nil {
		p.log.Warn().Err(err).Str("path", content.Path).
			Msg("failed to retrieve volume on disk")
		return false
	}

	if chapter.Volume != "" && onDiskVolume != chapter.Volume {
		p.log.Debug().Str("onDiskVolume", onDiskVolume).
			Str("volume", chapter.Volume).
			Msg("redownloading content")

		p.toRemoveContent = append(p.toRemoveContent, path.Join(p.client.GetBaseDir(), content.Path))
		return true
	}

	return false
}

func (p *publication) DownloadContent(ctx context.Context) {
	if p.state != payload.ContentStateReady && p.state != payload.ContentStateWaiting {
		p.log.Warn().Any("state", p.state).Msg("cannot start downloading in this state")
		return
	}

	p.SetState(payload.ContentStateDownloading)

	// Don't crash entire program when a download fails
	defer func() {
		if r := recover(); r != nil {
			p.log.Error().Any("error", r).Msg("recovering from panic during download")
		}
	}()

	p.startDownloadPipeline(ctx)
}

func (p *publication) GetNewContentNamed() []string {
	return utils.MaybeMap(p.toDownload, func(id string) (string, bool) {
		chapter, ok := p.getChapterById(id)
		return chapter.Label(), ok
	})
}

func (p *publication) FailedDownloads() int {
	return int(p.failedDownloads)
}

func (p *publication) UpdateSeriesInfo(f func(s *Series)) {
	f(p.series)
}
