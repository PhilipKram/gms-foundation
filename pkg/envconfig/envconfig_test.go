package envconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequired_Set(t *testing.T) {
	t.Setenv("TEST_REQUIRED", "myvalue")
	v, err := Required("TEST_REQUIRED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "myvalue" {
		t.Errorf("expected 'myvalue', got %q", v)
	}
}

func TestRequired_Missing(t *testing.T) {
	// Ensure the variable is unset
	t.Setenv("TEST_REQUIRED_MISSING", "")
	os.Unsetenv("TEST_REQUIRED_MISSING") //nolint:errcheck
	_, err := Required("TEST_REQUIRED_MISSING")
	if err == nil {
		t.Fatal("expected error for missing required variable")
	}
}

func TestRequired_Empty(t *testing.T) {
	t.Setenv("TEST_REQUIRED_EMPTY", "")
	_, err := Required("TEST_REQUIRED_EMPTY")
	if err == nil {
		t.Fatal("expected error for empty required variable")
	}
}

func TestOptional_Set(t *testing.T) {
	t.Setenv("TEST_OPT", "custom")
	if v := Optional("TEST_OPT", "default"); v != "custom" {
		t.Errorf("expected 'custom', got %q", v)
	}
}

func TestOptional_Missing(t *testing.T) {
	t.Setenv("TEST_OPT_MISSING", "")
	os.Unsetenv("TEST_OPT_MISSING") //nolint:errcheck
	if v := Optional("TEST_OPT_MISSING", "fallback"); v != "fallback" {
		t.Errorf("expected 'fallback', got %q", v)
	}
}

func TestOptionalBool_True(t *testing.T) {
	for _, val := range []string{"true", "1", "yes", "TRUE", "Yes"} {
		t.Setenv("TEST_BOOL", val)
		if !OptionalBool("TEST_BOOL", false) {
			t.Errorf("expected true for %q", val)
		}
	}
}

func TestOptionalBool_False(t *testing.T) {
	t.Setenv("TEST_BOOL_F", "false")
	if OptionalBool("TEST_BOOL_F", true) {
		t.Error("expected false for 'false'")
	}
}

func TestOptionalBool_Missing(t *testing.T) {
	t.Setenv("TEST_BOOL_MISSING", "")
	os.Unsetenv("TEST_BOOL_MISSING") //nolint:errcheck
	if OptionalBool("TEST_BOOL_MISSING", true) != true {
		t.Error("expected default true")
	}
	if OptionalBool("TEST_BOOL_MISSING", false) != false {
		t.Error("expected default false")
	}
}

func TestOptionalStringSlice_Set(t *testing.T) {
	t.Setenv("TEST_SLICE", "a, b , c")
	got := OptionalStringSlice("TEST_SLICE", ",", nil)
	if len(got) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(got))
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("expected [a b c], got %v", got)
	}
}

func TestOptionalStringSlice_Empty(t *testing.T) {
	os.Unsetenv("TEST_SLICE_EMPTY")
	got := OptionalStringSlice("TEST_SLICE_EMPTY", ",", []string{"default"})
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("expected [default], got %v", got)
	}
}

func TestOptionalStringSlice_AllWhitespace(t *testing.T) {
	t.Setenv("TEST_SLICE_WS", " , , ")
	got := OptionalStringSlice("TEST_SLICE_WS", ",", []string{"fallback"})
	if len(got) != 1 || got[0] != "fallback" {
		t.Errorf("expected [fallback] for all-whitespace input, got %v", got)
	}
}

func TestResolveAbsPath_Absolute(t *testing.T) {
	got, err := ResolveAbsPath("/usr/local/bin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/usr/local/bin" {
		t.Errorf("expected /usr/local/bin, got %q", got)
	}
}

func TestResolveAbsPath_Relative(t *testing.T) {
	got, err := ResolveAbsPath("./uploads")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %q", got)
	}
	// Should end with /uploads
	if filepath.Base(got) != "uploads" {
		t.Errorf("expected path ending in 'uploads', got %q", got)
	}
}
