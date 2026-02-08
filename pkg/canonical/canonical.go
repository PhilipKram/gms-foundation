package canonical

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
)

// trackingParams lists query parameters commonly used for tracking that should
// be stripped during URL canonicalization.
var trackingParams = map[string]bool{
	"utm_source":    true,
	"utm_medium":    true,
	"utm_campaign":  true,
	"utm_term":      true,
	"utm_content":   true,
	"utm_id":        true,
	"utm_cid":       true,
	"gclid":         true,
	"gclsrc":        true,
	"fbclid":        true,
	"dclid":         true,
	"msclkid":       true,
	"twclid":        true,
	"mc_cid":        true,
	"mc_eid":        true,
	"yclid":         true,
	"_ga":           true,
	"_gl":           true,
	"_hsenc":        true,
	"_hsmi":         true,
	"__hstc":        true,
	"__hssc":        true,
	"__hsfp":        true,
	"hsCtaTracking": true,
	"ref":           true,
	"ref_src":       true,
	"ref_url":       true,
	"source":        true,
	"si":            true,
}

var mu sync.RWMutex

// AddTrackingParams adds additional parameter names to the tracking-param strip
// list. This is safe for concurrent use.
func AddTrackingParams(params ...string) {
	mu.Lock()
	defer mu.Unlock()
	for _, p := range params {
		trackingParams[strings.ToLower(p)] = true
	}
}

// Canonicalize normalizes a raw URL into a canonical form and computes a
// SHA-256 hash of the result. The canonical form lowercases the scheme and host,
// removes fragments, strips tracking query parameters, sorts remaining query
// parameters, and normalizes trailing slashes.
func Canonicalize(rawURL string) (canonicalURL, hash string, err error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", "", fmt.Errorf("empty URL")
	}

	// Add scheme if missing
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("parsing URL: %w", err)
	}

	if parsed.Host == "" {
		return "", "", fmt.Errorf("URL has no host")
	}

	// Lowercase scheme and host
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	// Remove default ports
	host := parsed.Host
	if strings.HasSuffix(host, ":80") && parsed.Scheme == "http" {
		host = strings.TrimSuffix(host, ":80")
	}
	if strings.HasSuffix(host, ":443") && parsed.Scheme == "https" {
		host = strings.TrimSuffix(host, ":443")
	}
	parsed.Host = host

	// Remove fragment
	parsed.Fragment = ""
	parsed.RawFragment = ""

	// Normalize path: ensure single trailing slash for root, remove trailing
	// slash for other paths
	path := parsed.Path
	if path == "" {
		path = "/"
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	parsed.Path = path

	// Remove tracking params and sort remaining
	mu.RLock()
	query := parsed.Query()
	cleanQuery := url.Values{}
	for key, values := range query {
		lowerKey := strings.ToLower(key)
		if trackingParams[lowerKey] {
			continue
		}
		// Skip params that start with utm_
		if strings.HasPrefix(lowerKey, "utm_") {
			continue
		}
		for _, v := range values {
			cleanQuery.Add(key, v)
		}
	}
	mu.RUnlock()

	// Sort query parameters deterministically
	if len(cleanQuery) > 0 {
		keys := make([]string, 0, len(cleanQuery))
		for k := range cleanQuery {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var qParts []string
		for _, k := range keys {
			vals := cleanQuery[k]
			sort.Strings(vals)
			for _, v := range vals {
				qParts = append(qParts, url.QueryEscape(k)+"="+url.QueryEscape(v))
			}
		}
		parsed.RawQuery = strings.Join(qParts, "&")
	} else {
		parsed.RawQuery = ""
	}

	canonical := parsed.String()

	// Compute SHA-256 hash
	h := sha256.Sum256([]byte(canonical))
	hashStr := fmt.Sprintf("%x", h)

	return canonical, hashStr, nil
}
