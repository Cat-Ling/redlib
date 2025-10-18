package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDevice(t *testing.T) {
	device := newDevice()
	if device.UserAgent == "" {
		t.Errorf("Expected UserAgent to be non-empty")
	}
}

func TestNewOauth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"access_token": "test-token", "expires_in": 3600}`)
	}))
	defer server.Close()

	originalAuthEndpoint := authEndpoint
	authEndpoint = server.URL
	defer func() { authEndpoint = originalAuthEndpoint }()

	device := newDevice()
	oauth := newOauth(device)

	if oauth.Token != "test-token" {
		t.Errorf("Expected token to be 'test-token', got '%s'", oauth.Token)
	}
	if oauth.ExpiresIn != 3600 {
		t.Errorf("Expected ExpiresIn to be 3600, got %d", oauth.ExpiresIn)
	}
}