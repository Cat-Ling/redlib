package main

import (
	"testing"
)

func TestJSONRequest(t *testing.T) {
	// Using a public test API
	url := "https://jsonplaceholder.typicode.com/posts/1"
	resp, err := jsonRequest(url, false)
	if err != nil {
		t.Fatalf("jsonRequest failed: %v", err)
	}

	data, ok := resp.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected response to be a map, got %T", resp)
	}

	if _, ok := data["title"]; !ok {
		t.Errorf("Expected response to have a 'title' field")
	}
}