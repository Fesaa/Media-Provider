package services

import (
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"testing"
)

func inMemFs() afero.Afero {
	return afero.Afero{Fs: afero.NewMemMapFs()}
}

func mockDirectoryService(t *testing.T, fs afero.Afero) DirectoryService {
	t.Helper()
	return DirectoryServiceProvider(zerolog.Nop(), fs)
}

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestDirectoryService_ZipToCbz(t *testing.T) {
	fs := inMemFs()
	ds := mockDirectoryService(t, fs)

	must(t, fs.Mkdir("/data", 0755))
	must(t, fs.Mkdir("/out", 0755))
	_, err := fs.Create("/data/file1.txt")
	must(t, err)
	_, err = fs.Create("/data/file2.txt")
	must(t, err)
	_, err = fs.Create("/data/file3.txt")
	must(t, err)

	err = ds.ZipToCbz("/data")
	must(t, err)

	ok, err := fs.Exists("/data.cbz")
	must(t, err)
	if !ok {
		t.Fatal("expected zip file to exist")
	}
}

func TestDirectoryService_ZipFolder(t *testing.T) {
	fs := inMemFs()
	ds := mockDirectoryService(t, fs)

	must(t, fs.Mkdir("/data", 0755))
	must(t, fs.Mkdir("/out", 0755))
	_, err := fs.Create("/data/file1.txt")
	must(t, err)
	_, err = fs.Create("/data/file2.txt")
	must(t, err)
	_, err = fs.Create("/data/file3.txt")
	must(t, err)

	err = ds.ZipFolder("/data", "/out/myZip.zip")
	must(t, err)

	ok, err := fs.Exists("/out/myZip.zip")
	must(t, err)
	if !ok {
		t.Fatal("expected zip file to exist")
	}
}

func TestDirectoryService_ZipFolder_DirNotExists(t *testing.T) {
	fs := inMemFs()
	ds := mockDirectoryService(t, fs)

	err := ds.ZipFolder("DirNotFound", "out.zip")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDirectoryService_MoveDirectoryContent(t *testing.T) {
	fs := inMemFs()
	ds := mockDirectoryService(t, fs)

	must(t, fs.Mkdir("/data", 0755))
	_, err := fs.Create("/data/1.txt")
	must(t, err)
	_, err = fs.Create("/data/2.txt")
	must(t, err)

	must(t, ds.MoveDirectoryContent("/data", "/data2"))

	files, err := fs.ReadDir("/data2")
	must(t, err)

	want := 2
	if len(files) != want {
		t.Errorf("want %d files, got %d", want, len(files))
	}
}
