package mangadex

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

// ID https://mangadex.org/title/476c91e3-5eb2-438e-8aff-0ea3aae0479f/
// One chapter, with no volume. Very light for tests
const ID = "476c91e3-5eb2-438e-8aff-0ea3aae0479f"

// LONG_ID https://mangadex.org/title/8e74e420-b05e-4975-9844-676c7156bd63/
// 7 volumes, long for stop test
const LONG_ID = "8e74e420-b05e-4975-9844-676c7156bd63"

type mockConfig struct {
	d string
	m int
}

func (m *mockConfig) GetRootDir() string {
	return m.d
}

func (m *mockConfig) GetMaxConcurrentMangadexImages() int {
	return m.m
}

func TestQueue(t *testing.T) {
	mock := mockConfig{d: t.TempDir(), m: 4}
	client := newClient(&mock)

	var req = payload.DownloadRequest{
		Provider:  config.MANGADEX,
		Id:        LONG_ID,
		BaseDir:   "",
		TempTitle: "Love Affair Ranger",
	}
	var stopReq = payload.StopRequest{
		Provider:    config.MANGADEX,
		Id:          LONG_ID,
		DeleteFiles: true,
	}
	var req1 = payload.DownloadRequest{
		Provider:  config.MANGADEX,
		Id:        ID,
		BaseDir:   "",
		TempTitle: "Destiny Unchain Online",
	}
	var stopReq1 = payload.StopRequest{
		Provider:    config.MANGADEX,
		Id:          ID,
		DeleteFiles: true,
	}

	_, err := client.Download(req)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Download(req1)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	if len(client.GetQueuedMangas()) == 0 {
		t.Fatal(fmt.Errorf("queue was empty"))
	}

	_ = client.RemoveDownload(stopReq)
	time.Sleep(5 * time.Second)
	if len(client.GetQueuedMangas()) > 0 {
		t.Fatal(fmt.Errorf("queue not empty"))
	}
	_ = client.RemoveDownload(stopReq1)
	if client.GetCurrentManga() != nil {
		t.Fatal(fmt.Errorf("manga was not empty"))
	}
}

func TestCancel(t *testing.T) {
	mock := mockConfig{d: t.TempDir(), m: 4}
	client := newClient(&mock)

	var req = payload.DownloadRequest{
		Provider:  config.MANGADEX,
		Id:        LONG_ID,
		BaseDir:   "",
		TempTitle: "Destiny Unchain Online",
	}

	manga, err := client.Download(req)
	if err != nil {
		t.Fatal(err)
	}

	// Allow for download to start
	time.Sleep(5 * time.Second)
	stop := payload.StopRequest{
		Provider:    config.MANGADEX,
		Id:          LONG_ID,
		DeleteFiles: true,
	}

	err = client.RemoveDownload(stop)
	if err != nil {
		t.Fatal(err)
	}

	// Allow files to be deleted
	time.Sleep(5 * time.Second)

	dir := path.Join(mock.GetRootDir(), manga.Title())
	_, err = os.Stat(dir)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(fmt.Sprintf("directory '%s' was found. Should be gone", dir))
	}
}

func TestDownload(t *testing.T) {
	mock := mockConfig{d: t.TempDir(), m: 4}
	client := newClient(&mock)

	var req = payload.DownloadRequest{
		Provider:  config.MANGADEX,
		Id:        ID,
		BaseDir:   "",
		TempTitle: "Love Affair Ranger",
	}

	manga, err := client.Download(req)
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for range time.Tick(5 * time.Second) {
			if client.GetCurrentManga() == nil {
				wg.Done()
				break
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		time.Sleep(1 * time.Minute)
		wg.Done()
	}()

	wg.Wait()

	dir := path.Join(mock.GetRootDir(), manga.Title())
	stat, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !stat.IsDir() {
		t.Fatal(fmt.Errorf("expected a directory"))
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, entry := range dirEntries {
		if entry.Name() == "The Girlfriend I Care About Vol. .cbz" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Dowloaded file not found")
	}

}
