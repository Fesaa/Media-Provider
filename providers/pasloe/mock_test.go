package pasloe

import (
	"context"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"time"
)

type mockRegistry struct {
	cont          *dig.Container
	finishContent bool
}

func (m mockRegistry) Create(c core.Client, req payload.DownloadRequest) (core.Downloadable, error) {
	scope := m.cont.Scope("pasloe::registry::create")

	utils.Must(scope.Provide(utils.Identity(c)))
	utils.Must(scope.Provide(utils.Identity(req)))

	block := &MockContent{
		mockId:       req.Id,
		mockProvider: req.Provider,
	}

	if !m.finishContent {
		block.mockDownloadContentFunc = func(idx int, t ID, url string) error {
			time.Sleep(time.Millisecond * 100)
			return nil
		}
		block.mockContentUrlsFunc = func(ctx context.Context, t ID) ([]string, error) {
			return []string{"a", "b", "c", "d", "e"}, nil
		}
		block.mockAll = []ID{"a", "b"}
	}

	base := core.New[ID, ID](scope, "", block)
	block.Core = base
	return block, nil
}

type ID string

func (i ID) AllChapters() []ID {
	return []ID{}
}

func (i ID) GetChapter() string {
	return i.GetId()
}

func (i ID) GetVolume() string {
	return i.GetId()
}

func (i ID) GetTitle() string {
	return i.GetId()
}

func (i ID) GetId() string {
	return string(i)
}

func (i ID) Label() string {
	return string(i)
}

type MockContent struct {
	*core.Core[ID, ID]
	mockTitle               string
	mockRefUrl              string
	mockProvider            models.Provider
	mockInfo                payload.InfoStat
	mockId                  string
	mockAll                 []ID
	mockContentList         []payload.ListContentData
	mockContentDirFunc      func(t ID) string
	mockContentPathFunc     func(t ID) string
	mockContentKeyFunc      func(t ID) string
	mockContentLoggerFunc   func(t ID) zerolog.Logger
	mockContentUrlsFunc     func(ctx context.Context, t ID) ([]string, error)
	mockWriteMetaDataFunc   func(t ID) error
	mockDownloadContentFunc func(idx int, t ID, url string) error
	mockIsContentFunc       func(s string) bool
	mockShouldDownloadFunc  func(t ID) bool
	loadInfoFunc            func()
	loadInfoChan            chan struct{}
}

func NewMockContent(scope *dig.Scope) *MockContent {
	mc := &MockContent{}
	base := core.New[ID, ID](scope, "mock-content", mc)
	mc.Core = base
	return mc
}

func (m *MockContent) Id() string {
	return m.mockId
}

func (m *MockContent) Title() string {
	return m.mockTitle
}

func (m *MockContent) RefUrl() string {
	return m.mockRefUrl
}

func (m *MockContent) Provider() models.Provider {
	return m.mockProvider
}

func (m *MockContent) LoadInfo(ctx context.Context) chan struct{} {
	m.loadInfoChan = make(chan struct{})
	go func() {
		if m.loadInfoFunc != nil {
			m.loadInfoFunc()
		} else {
			close(m.loadInfoChan)
		}
	}()

	return m.loadInfoChan
}

func (m *MockContent) GetInfo() payload.InfoStat {
	return m.mockInfo
}

func (m *MockContent) CustomizeAllChapters() []ID {
	return m.mockAll
}

func (m *MockContent) ContentList() []payload.ListContentData {
	return m.mockContentList
}

func (m *MockContent) ContentDir(t ID) string {
	if m.mockContentDirFunc != nil {
		return m.mockContentDirFunc(t)
	}
	return t.GetId()
}

func (m *MockContent) ContentPath(t ID) string {
	if m.mockContentPathFunc != nil {
		return m.mockContentPathFunc(t)
	}
	return t.GetId()
}

func (m *MockContent) ContentKey(t ID) string {
	if m.mockContentKeyFunc != nil {
		return m.mockContentKeyFunc(t)
	}
	return t.GetId()
}

func (m *MockContent) ContentLogger(t ID) zerolog.Logger {
	if m.mockContentLoggerFunc != nil {
		return m.mockContentLoggerFunc(t)
	}
	return zerolog.Nop()
}

func (m *MockContent) ContentUrls(ctx context.Context, t ID) ([]string, error) {
	if m.mockContentUrlsFunc != nil {
		return m.mockContentUrlsFunc(ctx, t)
	}
	return nil, nil
}

func (m *MockContent) WriteContentMetaData(t ID) error {
	if m.mockWriteMetaDataFunc != nil {
		return m.mockWriteMetaDataFunc(t)
	}
	return nil
}

func (m *MockContent) DownloadContent(idx int, t ID, url string) error {
	if m.mockDownloadContentFunc != nil {
		return m.mockDownloadContentFunc(idx, t, url)
	}
	return nil
}

func (m *MockContent) IsContent(s string) bool {
	if m.mockIsContentFunc != nil {
		return m.mockIsContentFunc(s)
	}
	return true
}

func (m *MockContent) ShouldDownload(t ID) bool {
	if m.mockShouldDownloadFunc != nil {
		return m.mockShouldDownloadFunc(t)
	}
	return true
}
