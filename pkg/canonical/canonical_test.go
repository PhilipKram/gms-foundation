package canonical

import (
	"testing"
)

func TestCanonicalize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "basic normalization",
			input: "HTTPS://Example.COM/path",
			want:  "https://example.com/path",
		},
		{
			name:  "strips utm params",
			input: "https://example.com/page?utm_source=twitter&utm_medium=social&keep=yes",
			want:  "https://example.com/page?keep=yes",
		},
		{
			name:  "strips fbclid",
			input: "https://example.com/?fbclid=abc123&q=test",
			want:  "https://example.com/?q=test",
		},
		{
			name:  "sorts query params",
			input: "https://example.com/?z=1&a=2&m=3",
			want:  "https://example.com/?a=2&m=3&z=1",
		},
		{
			name:  "removes fragment",
			input: "https://example.com/page#section",
			want:  "https://example.com/page",
		},
		{
			name:  "removes default http port",
			input: "http://example.com:80/path",
			want:  "http://example.com/path",
		},
		{
			name:  "removes default https port",
			input: "https://example.com:443/path",
			want:  "https://example.com/path",
		},
		{
			name:  "keeps non-default port",
			input: "https://example.com:8080/path",
			want:  "https://example.com:8080/path",
		},
		{
			name:  "adds https scheme if missing",
			input: "example.com/path",
			want:  "https://example.com/path",
		},
		{
			name:  "root path trailing slash",
			input: "https://example.com",
			want:  "https://example.com/",
		},
		{
			name:  "removes trailing slash on non-root",
			input: "https://example.com/path/",
			want:  "https://example.com/path",
		},
		{
			name:  "strips custom utm prefix",
			input: "https://example.com/?utm_custom_thing=x&q=1",
			want:  "https://example.com/?q=1",
		},
		{
			name:    "empty URL",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:  "strips all tracking leaves no query",
			input: "https://example.com/page?utm_source=a&fbclid=b&gclid=c",
			want:  "https://example.com/page",
		},
		{
			name:  "preserves non-tracking params",
			input: "https://example.com/search?q=hello+world&lang=en",
			want:  "https://example.com/search?lang=en&q=hello+world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, hash, err := Canonicalize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Canonicalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if hash == "" {
				t.Error("expected non-empty hash")
			}
		})
	}
}

func TestCanonicalize_NoHost(t *testing.T) {
	_, _, err := Canonicalize("https:///path/only")
	if err == nil {
		t.Fatal("expected error for URL with no host")
	}
}

func TestCanonicalize_MultipleValuesForSameKey(t *testing.T) {
	got, _, err := Canonicalize("https://example.com/?a=2&a=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Values for same key should be sorted
	want := "https://example.com/?a=1&a=2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCanonicalize_HttpScheme(t *testing.T) {
	got, _, err := Canonicalize("HTTP://Example.COM/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://example.com/path" {
		t.Errorf("got %q, want %q", got, "http://example.com/path")
	}
}

func TestCanonicalize_AllKnownTrackingParams(t *testing.T) {
	// Test a selection of known tracking params that aren't utm_*
	params := []string{"gclid", "fbclid", "msclkid", "twclid", "_ga", "_gl", "si", "mc_cid", "yclid"}
	for _, p := range params {
		got, _, err := Canonicalize("https://example.com/?" + p + "=val&keep=1")
		if err != nil {
			t.Fatalf("unexpected error for param %s: %v", p, err)
		}
		if got != "https://example.com/?keep=1" {
			t.Errorf("param %s not stripped: got %q", p, got)
		}
	}
}

func TestCanonicalize_HashConsistency(t *testing.T) {
	_, h1, _ := Canonicalize("https://example.com/page?b=2&a=1")
	_, h2, _ := Canonicalize("https://EXAMPLE.COM/page?a=1&b=2")
	if h1 != h2 {
		t.Errorf("same canonical URL should produce same hash: %q != %q", h1, h2)
	}
}

func TestCanonicalize_DifferentURLsDifferentHashes(t *testing.T) {
	_, h1, _ := Canonicalize("https://example.com/page1")
	_, h2, _ := Canonicalize("https://example.com/page2")
	if h1 == h2 {
		t.Error("different URLs should produce different hashes")
	}
}

func TestAddTrackingParams(t *testing.T) {
	AddTrackingParams("my_tracker", "custom_id")

	got, _, err := Canonicalize("https://example.com/?my_tracker=x&custom_id=y&keep=z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.com/?keep=z"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
