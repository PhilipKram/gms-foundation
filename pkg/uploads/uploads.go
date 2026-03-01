package uploads

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// FileCategory defines allowed MIME types, size limits, and storage subdirectory
// for a category of uploads.
type FileCategory struct {
	Subdir       string
	MaxSize      int64
	AllowedTypes map[string]string // MIME type -> file extension
}

// DefaultImageCategory returns a FileCategory for common image types (10 MB max).
func DefaultImageCategory() FileCategory {
	return FileCategory{
		Subdir:  "images",
		MaxSize: 10 << 20,
		AllowedTypes: map[string]string{
			"image/jpeg": ".jpg",
			"image/png":  ".png",
			"image/gif":  ".gif",
			"image/webp": ".webp",
		},
	}
}

// DefaultAudioCategory returns a FileCategory for common audio types (50 MB max).
func DefaultAudioCategory() FileCategory {
	return FileCategory{
		Subdir:  "audio",
		MaxSize: 50 << 20,
		AllowedTypes: map[string]string{
			"audio/mpeg":  ".mp3",
			"audio/wav":   ".wav",
			"audio/x-m4a": ".m4a",
			"audio/mp4":   ".m4a",
			"audio/ogg":   ".ogg",
		},
	}
}

type config struct {
	categories []FileCategory
}

// Option configures a Storage instance.
type Option func(*config)

// WithCategories replaces the default categories with the provided ones.
func WithCategories(cats ...FileCategory) Option {
	return func(c *config) { c.categories = cats }
}

// WithCategory appends a category to the list.
func WithCategory(cat FileCategory) Option {
	return func(c *config) { c.categories = append(c.categories, cat) }
}

// Storage handles file persistence for uploaded files.
type Storage struct {
	baseDir    string
	categories []FileCategory
}

// NewStorage creates a Storage and ensures all category subdirectories exist.
// If no options are provided, DefaultImageCategory and DefaultAudioCategory are used.
func NewStorage(baseDir string, opts ...Option) (*Storage, error) {
	cfg := config{
		categories: []FileCategory{DefaultImageCategory(), DefaultAudioCategory()},
	}
	for _, o := range opts {
		o(&cfg)
	}

	for _, cat := range cfg.categories {
		dir := filepath.Join(baseDir, cat.Subdir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating upload directory %s: %w", dir, err)
		}
	}

	return &Storage{baseDir: baseDir, categories: cfg.categories}, nil
}

// SaveFile validates the MIME type against configured categories, generates a
// unique filename, validates the content via magic-byte detection, and writes
// the file to disk. It returns the relative path (e.g. "images/abc.jpg").
func (s *Storage) SaveFile(r io.Reader, mimeType string) (string, error) {
	cat, ext, err := s.classify(mimeType)
	if err != nil {
		return "", err
	}

	// Read into memory-limited buffer to enforce size limits.
	limited := io.LimitReader(r, cat.MaxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("reading upload: %w", err)
	}
	if int64(len(data)) > cat.MaxSize {
		return "", fmt.Errorf("file exceeds maximum size of %d bytes", cat.MaxSize)
	}

	// Validate that file content matches the declared MIME type.
	if err := validateContent(data, mimeType, cat); err != nil {
		return "", err
	}

	filename := uuid.New().String() + ext
	relPath := filepath.Join(cat.Subdir, filename)
	absPath := filepath.Join(s.baseDir, relPath)

	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return relPath, nil
}

// DeleteFile removes a file by its relative path, with path traversal protection.
func (s *Storage) DeleteFile(relPath string) error {
	cleaned := filepath.Clean(relPath)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	absPath := filepath.Join(s.baseDir, cleaned)
	resolved, err := filepath.Abs(absPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}
	base, err := filepath.Abs(s.baseDir)
	if err != nil {
		return fmt.Errorf("resolving base: %w", err)
	}

	// Use filepath.Rel for robust containment check.
	rel, err := filepath.Rel(base, resolved)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("invalid file path: outside base directory")
	}

	return os.Remove(resolved)
}

// CategoryFor returns the subdirectory name (e.g. "images", "audio") for a
// given MIME type, or an empty string if the type is not recognized.
func (s *Storage) CategoryFor(mimeType string) string {
	for _, cat := range s.categories {
		if _, ok := cat.AllowedTypes[mimeType]; ok {
			return cat.Subdir
		}
	}
	return ""
}

func (s *Storage) classify(mimeType string) (FileCategory, string, error) {
	for _, cat := range s.categories {
		if ext, ok := cat.AllowedTypes[mimeType]; ok {
			return cat, ext, nil
		}
	}
	return FileCategory{}, "", fmt.Errorf("file type %q not supported", mimeType)
}

// validateContent checks that actual file content matches the declared MIME type
// using magic-byte detection.
func validateContent(data []byte, declaredType string, cat FileCategory) error {
	if len(data) == 0 {
		return fmt.Errorf("empty file")
	}

	detected := http.DetectContentType(data)

	// For image categories, detected type must exactly match the declared type.
	if strings.HasPrefix(declaredType, "image/") {
		if detected != declaredType {
			return fmt.Errorf("file content type %q does not match declared type %q", detected, declaredType)
		}
		return nil
	}

	// For audio, magic-byte detection is less reliable (e.g. m4a may detect as
	// video/mp4 or application/octet-stream). Use an allowlist of detected
	// content types that are known to correspond to valid audio content.
	if strings.HasPrefix(declaredType, "audio/") {
		switch {
		case strings.HasPrefix(detected, "audio/"): // e.g. audio/mpeg for ID3-tagged MP3
		case detected == "application/octet-stream": // common for WAV, OGG, raw MP3
		case detected == "video/mp4": // M4A files detected as video/mp4
		default:
			return fmt.Errorf("file content type %q does not match declared audio type %q", detected, declaredType)
		}
		return nil
	}

	// For unknown MIME prefixes, allow through — the classify step already
	// validated that the type is in the category's AllowedTypes map.
	return nil
}
