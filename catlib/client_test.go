package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const popularURL = "/r/popular/hot.json?&raw_json=1&geo_filter=GLOBAL"

func TestRateLimitCheck(t *testing.T) {
	// This test requires a live connection to Reddit and valid OAuth credentials.
	// It's more of an integration test.
	t.Skip("Skipping rate limit check test due to external dependency.")
	if err := rateLimitCheck(); err != nil {
		t.Errorf("rateLimitCheck() failed: %v", err)
	}
}

func TestLocalizationPopular(t *testing.T) {
	// This test also requires a live connection.
	t.Skip("Skipping localization popular test due to external dependency.")
	val, err := jsonRequest(popularURL, false)
	if err != nil {
		t.Fatalf("jsonRequest failed: %v", err)
	}

	data, ok := val.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected response to be a map, got %T", val)
	}

	geoFilter := data["data"].(map[string]interface{})["geo_filter"].(string)
	if geoFilter != "GLOBAL" {
		t.Errorf("Expected geo_filter to be 'GLOBAL', got '%s'", geoFilter)
	}
}

func TestObfuscatedShareLink(t *testing.T) {
	// Mocking the Reddit API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/r/rust/s/kPgq8WNHRK") {
			w.Header().Set("Location", "/r/rust/comments/18t5968/why_use_tuple_struct_over_standard_struct/kfbqlbc/.json")
			w.WriteHeader(http.StatusMovedPermanently)
		} else if strings.HasPrefix(r.URL.Path, "/r/rust/comments/18t5968/why_use_tuple_struct_over_standard_struct/kfbqlbc/") {
			w.WriteHeader(http.StatusOK)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Temporarily replace the base URLs to point to the mock server
	mockRedditShortURLBase := server.URL
	originalURLPairs := urlPairs
	urlPairs = []struct {
		Base string
		Host string
	}{
		{alternativeRedditURLBase, alternativeRedditURLBaseHost},
		{mockRedditShortURLBase, ""},
	}
	defer func() { urlPairs = originalURLPairs }()

	shareLink := "/r/rust/s/kPgq8WNHRK"
	canonicalLink := "/r/rust/comments/18t5968/why_use_tuple_struct_over_standard_struct/kfbqlbc/"

	path, err := canonicalPath(shareLink, 3)
	if err != nil {
		t.Fatalf("canonicalPath failed: %v", err)
	}

	if path != canonicalLink {
		t.Errorf("Expected canonical path '%s', got '%s'", canonicalLink, path)
	}
}

func TestPrivateSub(t *testing.T) {
	t.Skip("Skipping private sub test due to external dependency.")
	_, err := jsonRequest("/r/suicide/about.json?raw_json=1", true)
	if err == nil || err.Error() != "private" {
		t.Errorf("Expected 'private' error, got %v", err)
	}
}

func TestBannedSub(t *testing.T) {
	t.Skip("Skipping banned sub test due to external dependency.")
	_, err := jsonRequest("/r/aaa/about.json?raw_json=1", true)
	if err == nil || err.Error() != "banned" {
		t.Errorf("Expected 'banned' error, got %v", err)
	}
}

func TestGatedSub(t *testing.T) {
	t.Skip("Skipping gated sub test due to external dependency.")
	_, err := jsonRequest("/r/drugs/about.json?raw_json=1", false)
	if err == nil || err.Error() != "gated" {
		t.Errorf("Expected 'gated' error, got %v", err)
	}
}
