package uploads

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalJPEG is a valid JPEG file header (SOI + APP0 JFIF marker).
var minimalJPEG = []byte{
	0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
	0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
	0x00, 0x01, 0x00, 0x00,
}

// minimalPNG is a valid PNG file signature.
var minimalPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
	0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
	0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
	0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestNewStorage_CreatesSubdirs(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = s

	for _, sub := range []string{"images", "audio"} {
		info, err := os.Stat(filepath.Join(dir, sub))
		if err != nil {
			t.Errorf("expected %s directory to exist: %v", sub, err)
		} else if !info.IsDir() {
			t.Errorf("expected %s to be a directory", sub)
		}
	}
}

func TestNewStorage_CustomCategories(t *testing.T) {
	dir := t.TempDir()
	cat := FileCategory{
		Subdir:       "docs",
		MaxSize:      1 << 20,
		AllowedTypes: map[string]string{"application/pdf": ".pdf"},
	}
	s, err := NewStorage(dir, WithCategories(cat))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "docs"))
	if err != nil {
		t.Fatalf("expected docs directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected docs to be a directory")
	}

	// Default categories should not exist
	if _, err := os.Stat(filepath.Join(dir, "images")); !os.IsNotExist(err) {
		t.Error("images directory should not exist with custom categories")
	}
	_ = s
}

func TestSaveFile_JPEG(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	relPath, err := s.SaveFile(bytes.NewReader(minimalJPEG), "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(relPath, "images/") {
		t.Errorf("expected images/ prefix, got %q", relPath)
	}
	if !strings.HasSuffix(relPath, ".jpg") {
		t.Errorf("expected .jpg suffix, got %q", relPath)
	}

	// File should exist on disk
	if _, err := os.Stat(filepath.Join(dir, relPath)); err != nil {
		t.Errorf("saved file should exist: %v", err)
	}
}

func TestSaveFile_PNG(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	relPath, err := s.SaveFile(bytes.NewReader(minimalPNG), "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(relPath, ".png") {
		t.Errorf("expected .png suffix, got %q", relPath)
	}
}

func TestSaveFile_UnsupportedType(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	_, err := s.SaveFile(bytes.NewReader([]byte("hello")), "text/plain")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestSaveFile_ExceedsMaxSize(t *testing.T) {
	dir := t.TempDir()
	tiny := FileCategory{
		Subdir:       "tiny",
		MaxSize:      10,
		AllowedTypes: map[string]string{"image/jpeg": ".jpg"},
	}
	s, _ := NewStorage(dir, WithCategories(tiny))

	data := make([]byte, 20)
	copy(data, minimalJPEG[:8]) // JPEG header
	_, err := s.SaveFile(bytes.NewReader(data), "image/jpeg")
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("expected size error, got: %v", err)
	}
}

func TestSaveFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	_, err := s.SaveFile(bytes.NewReader([]byte{}), "image/jpeg")
	if err == nil {
		t.Fatal("expected error for empty file")
	}
	if !strings.Contains(err.Error(), "empty file") {
		t.Errorf("expected empty file error, got: %v", err)
	}
}

func TestDeleteFile(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	relPath, _ := s.SaveFile(bytes.NewReader(minimalJPEG), "image/jpeg")

	if err := s.DeleteFile(relPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, relPath)); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestDeleteFile_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	err := s.DeleteFile("../../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Errorf("expected path traversal error, got: %v", err)
	}
}

func TestCategoryFor(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	if got := s.CategoryFor("image/jpeg"); got != "images" {
		t.Errorf("expected 'images', got %q", got)
	}
	if got := s.CategoryFor("audio/mpeg"); got != "audio" {
		t.Errorf("expected 'audio', got %q", got)
	}
	if got := s.CategoryFor("text/plain"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestWithCategory_Appends(t *testing.T) {
	dir := t.TempDir()
	extra := FileCategory{
		Subdir:       "videos",
		MaxSize:      100 << 20,
		AllowedTypes: map[string]string{"video/mp4": ".mp4"},
	}
	s, err := NewStorage(dir, WithCategory(extra))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have images, audio, and videos
	if got := s.CategoryFor("image/jpeg"); got != "images" {
		t.Errorf("expected 'images', got %q", got)
	}
	if got := s.CategoryFor("video/mp4"); got != "videos" {
		t.Errorf("expected 'videos', got %q", got)
	}
}
