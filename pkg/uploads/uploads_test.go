package uploads

import (
	"bytes"
	"fmt"
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

// errorReader always returns an error on Read.
type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("simulated read error")
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

func TestSaveFile_Audio(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	// MP3 files start with ID3 tag or 0xFF 0xFB sync bytes.
	// http.DetectContentType returns "application/octet-stream" for most audio,
	// which is fine — audio validation only rejects clearly wrong categories.
	mp3Header := []byte("ID3\x04\x00\x00\x00\x00\x00\x00")
	mp3Data := make([]byte, 100)
	copy(mp3Data, mp3Header)

	relPath, err := s.SaveFile(bytes.NewReader(mp3Data), "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(relPath, "audio/") {
		t.Errorf("expected audio/ prefix, got %q", relPath)
	}
	if !strings.HasSuffix(relPath, ".mp3") {
		t.Errorf("expected .mp3 suffix, got %q", relPath)
	}
}

func TestSaveFile_AudioContentMismatch(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	// Declare audio but send image content — should be rejected
	_, err := s.SaveFile(bytes.NewReader(minimalJPEG), "audio/mpeg")
	if err == nil {
		t.Fatal("expected error for audio content mismatch")
	}
	if !strings.Contains(err.Error(), "does not match declared audio type") {
		t.Errorf("expected audio mismatch error, got: %v", err)
	}
}

func TestSaveFile_ImageContentMismatch(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	// Declare image/jpeg but send non-image content
	plainText := []byte("this is not an image at all, just plain text data here")
	_, err := s.SaveFile(bytes.NewReader(plainText), "image/jpeg")
	if err == nil {
		t.Fatal("expected error for image content mismatch")
	}
	if !strings.Contains(err.Error(), "does not match declared image type") {
		t.Errorf("expected image mismatch error, got: %v", err)
	}
}

func TestDeleteFile_NonExistent(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	err := s.DeleteFile("images/nonexistent.jpg")
	if err == nil {
		t.Fatal("expected error for deleting non-existent file")
	}
}

func TestSaveFile_ReadError(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	_, err := s.SaveFile(&errorReader{}, "image/jpeg")
	if err == nil {
		t.Fatal("expected error from read failure")
	}
	if !strings.Contains(err.Error(), "reading upload") {
		t.Errorf("expected reading error, got: %v", err)
	}
}

func TestDefaultImageCategory(t *testing.T) {
	cat := DefaultImageCategory()
	if cat.Subdir != "images" {
		t.Errorf("expected subdir 'images', got %q", cat.Subdir)
	}
	if cat.MaxSize != 10<<20 {
		t.Errorf("expected max size 10MB, got %d", cat.MaxSize)
	}
	if len(cat.AllowedTypes) != 4 {
		t.Errorf("expected 4 allowed types, got %d", len(cat.AllowedTypes))
	}
	for _, mime := range []string{"image/jpeg", "image/png", "image/gif", "image/webp"} {
		if _, ok := cat.AllowedTypes[mime]; !ok {
			t.Errorf("expected %s in allowed types", mime)
		}
	}
}

func TestDefaultAudioCategory(t *testing.T) {
	cat := DefaultAudioCategory()
	if cat.Subdir != "audio" {
		t.Errorf("expected subdir 'audio', got %q", cat.Subdir)
	}
	if cat.MaxSize != 50<<20 {
		t.Errorf("expected max size 50MB, got %d", cat.MaxSize)
	}
	if len(cat.AllowedTypes) != 5 {
		t.Errorf("expected 5 allowed types, got %d", len(cat.AllowedTypes))
	}
}

func TestSaveFile_UniqueFilenames(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	p1, _ := s.SaveFile(bytes.NewReader(minimalJPEG), "image/jpeg")
	p2, _ := s.SaveFile(bytes.NewReader(minimalJPEG), "image/jpeg")
	if p1 == p2 {
		t.Error("expected unique filenames for each save")
	}
}

func TestSaveFile_WriteFailure(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	// Remove the images subdir to cause os.WriteFile to fail
	if err := os.Remove(filepath.Join(dir, "images")); err != nil { //nolint:errcheck
		t.Skipf("could not remove images dir: %v", err)
	}

	_, err := s.SaveFile(bytes.NewReader(minimalJPEG), "image/jpeg")
	if err == nil {
		t.Fatal("expected error when write directory is missing")
	}
	if !strings.Contains(err.Error(), "writing file") {
		t.Errorf("expected 'writing file' error, got: %v", err)
	}
}

func TestDeleteFile_CleanPath(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	// Save a file, then delete it using a path with extra slashes
	relPath, _ := s.SaveFile(bytes.NewReader(minimalJPEG), "image/jpeg")
	// Add redundant slashes — filepath.Clean should normalize
	messyPath := "images//" + filepath.Base(relPath)
	if err := s.DeleteFile(messyPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCategoryFor_AllDefaultTypes(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStorage(dir)

	imageMimes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	for _, m := range imageMimes {
		if got := s.CategoryFor(m); got != "images" {
			t.Errorf("CategoryFor(%q) = %q, want 'images'", m, got)
		}
	}

	audioMimes := []string{"audio/mpeg", "audio/wav", "audio/x-m4a", "audio/mp4", "audio/ogg"}
	for _, m := range audioMimes {
		if got := s.CategoryFor(m); got != "audio" {
			t.Errorf("CategoryFor(%q) = %q, want 'audio'", m, got)
		}
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
