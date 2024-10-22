package yoitsu

import (
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

const INFO_HASH_SMALL = "7edf3e2321e53044118e63ed255cb7834c264363"
const INFO_HASH_LARGE = "8d9686c1ac7beb16ca49c82ff5fde5f2ec7077e9"

type mockConfig struct {
	d string
	m int
}

func (m *mockConfig) GetRootDir() string {
	return m.d
}

func (m *mockConfig) GetMaxConcurrentTorrents() int {
	return m.m
}

func TestCancel(t *testing.T) {
	t.Parallel()
	mock := mockConfig{d: t.TempDir(), m: 5}
	client, err := newYoitsu(&mock)
	if err != nil {
		t.Fatal(err)
	}

	var req = payload.DownloadRequest{
		Provider:  config.NYAA,
		Id:        INFO_HASH_LARGE,
		BaseDir:   "",
		TempTitle: "[LonelyChaser] Getter Robo Go - 29 [CB6974A2] ",
	}

	tor, err := client.AddDownload(req)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	stop := payload.StopRequest{
		Provider:    config.NYAA,
		Id:          INFO_HASH_LARGE,
		DeleteFiles: true,
	}

	err = client.RemoveDownload(stop)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	dir := path.Join(mock.GetRootDir(), tor.GetTorrent().Name())
	dir2 := path.Join(mock.GetRootDir(), tor.GetTorrent().InfoHash().HexString())
	_, err = os.Stat(dir)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("expected ErrNotExist error, got " + func() string {
			if err != nil {
				return err.Error()
			}
			return "<nil>"
		}())
	}
	_, err = os.Stat(dir2)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("expected ErrNotExist error, got " + func() string {
			if err != nil {
				return err.Error()
			}
			return "<nil>"
		}())
	}

}

func TestDownload(t *testing.T) {
	t.Parallel()
	mock := mockConfig{d: t.TempDir(), m: 5}
	client, err := newYoitsu(&mock)
	if err != nil {
		t.Fatal(err)
	}

	var req = payload.DownloadRequest{
		Provider:  config.NYAA,
		Id:        INFO_HASH_SMALL,
		BaseDir:   "",
		TempTitle: "Classroom of the Elite - Year 2 - Volume 09 [Seven Seas]",
	}

	tor, err := client.AddDownload(req)
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for range time.Tick(5 * time.Second) {
			if client.GetRunningTorrents().Len() == 0 {
				wg.Done()
				break
			}
		}
	}()

	go func() {
		time.Sleep(1 * time.Minute)
		wg.Done()
	}()

	wg.Wait()

	dir := path.Join(mock.GetRootDir(), tor.GetTorrent().Name())
	stat, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !stat.IsDir() {
		t.Fatal("expected directory, got file")
	}

	found := false
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range dirEntries {
		if entry.Name() == "Classroom of the Elite - Year 2 - Volume 09.epub" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected file, got none")
	}

}
