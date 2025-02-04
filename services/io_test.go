package services

import (
	"archive/zip"
	"github.com/rs/zerolog"
	"io"
	"os"
	"path/filepath"
	"testing"
)

//nolint:funlen
func TestZipFolder(t *testing.T) {
	ios := IOServiceProvider(zerolog.Nop())
	tempDir := t.TempDir()

	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "file2.txt")
	subDir := filepath.Join(tempDir, "subdir")
	subFile := filepath.Join(subDir, "file3.txt")

	if err := os.WriteFile(testFile1, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("Go is awesome!"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(subFile, []byte("Inside subdirectory"), 0644); err != nil {
		t.Fatalf("Failed to create test file in subdirectory: %v", err)
	}

	zipFile := filepath.Join(tempDir, "test.zip")

	if err := ios.ZipFolder(tempDir, zipFile); err != nil {
		t.Fatalf("ZipFolder failed: %v", err)
	}

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		t.Fatalf("Failed to open zip file: %v", err)
	}
	defer r.Close()

	expectedFiles := map[string]string{
		"file1.txt":        "Hello, World!",
		"file2.txt":        "Go is awesome!",
		"subdir/file3.txt": "Inside subdirectory",
	}

	for _, f := range r.File {
		content, exists := expectedFiles[f.Name]
		if !exists {
			if f.Name != "test.zip" {
				t.Errorf("Unexpected file in zip: %s", f.Name)
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Fatalf("Failed to open file in zip: %v", err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("Failed to read file in zip: %v", err)
		}

		if string(data) != content {
			t.Errorf("File content mismatch for %s: expected %q, got %q", f.Name, content, string(data))
		}

		delete(expectedFiles, f.Name)
	}

	if len(expectedFiles) > 0 {
		t.Errorf("Missing files in zip: %v", expectedFiles)
	}
}

func TestZipInvalidFolder(t *testing.T) {
	ios := IOServiceProvider(zerolog.Nop())
	err := ios.ZipFolder("SRDETCFYVGUBHINJK", "EDTRFYGUJIK.zip")
	if err == nil {
		t.Errorf("ZipFolder should have failed on invalid folder")
	}
}

func TestZipInvalidFile(t *testing.T) {
	ios := IOServiceProvider(zerolog.Nop())
	tempDir := t.TempDir()
	testFile1 := filepath.Join(tempDir, "////")

	err := ios.ZipFolder(tempDir, testFile1)
	if err == nil {
		t.Errorf("ZipFolder should have failed on invalid file")
	}
}
