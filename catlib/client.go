package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

const (
	redditURLBase               = "https://oauth.reddit.com"
	redditURLBaseHost           = "oauth.reddit.com"
	redditShortURLBase          = "https://redd.it"
	redditShortURLBaseHost      = "redd.it"
	alternativeRedditURLBase    = "https://www.reddit.com"
	alternativeRedditURLBaseHost = "www.reddit.com"
)

var (
	httpClient *http.Client
	device     atomic.Value // Holds a Device
	oauthClient atomic.Value // Holds an Oauth

	oauthRatelimitRemaining uint32 = 99
	oauthIsRollingOver      atomic.Bool

	urlPairs = []struct {
		Base string
		Host string
	}{
		{alternativeRedditURLBase, alternativeRedditURLBaseHost},
		{redditShortURLBase, redditShortURLBaseHost},
	}
)


func init() {
	transport := &http.Transport{
		// from hyper_rustls::HttpsConnectorBuilder::new().with_native_roots().https_only().enable_http2().build()
		// Go's default transport has reasonable defaults that cover this.
		// We can customize it further if needed.
	}
	httpClient = &http.Client{
		Transport: transport,
		// Go's client handles redirects by default, which we need to disable for some requests.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

    // Initialize placeholder device and oauth client
    device.Store(Device{UserAgent: "catlib-go/0.1"})
    oauthClient.Store(Oauth{HeadersMap: make(map[string]string)})
}

// canonicalPath gets the canonical path for a resource on Reddit.
func canonicalPath(path string, tries int) (string, error) {
	if tries == 0 {
		return "", nil // Ok(None)
	}

	var finalResp *http.Response
	var err error

	for _, pair := range urlPairs {
		resp, err := redditShortHead(path, true, pair.Base, pair.Host)
		if err == nil {
			if resp.StatusCode < 400 {
				finalResp = resp
				break
			}
			resp.Body.Close()
		}
	}

	if finalResp == nil {
		return "", fmt.Errorf("unable to make HEAD request to Reddit: %v", err)
	}
	defer finalResp.Body.Close()

	status := finalResp.StatusCode

	if finalResp.Header.Get("Retry-After") != "" {
		return "", fmt.Errorf("too many requests")
	}

	switch {
	case status >= 200 && status <= 299:
		return path, nil
	case status == 301:
		location := finalResp.Header.Get("Location")
		if location == "" {
			return "", nil // Ok(None)
		}
		// Strip .json suffix and query parameters
		strippedURI := strings.Split(strings.TrimSuffix(location, ".json"), "?")[0]
		uri := formatURL(strippedURI)
		return canonicalPath(uri, tries-1)
	case status >= 300 && status <= 399:
		return "", nil // Ok(None)
	case status == 429:
		return "", fmt.Errorf("too many requests")
	case status == 403 && finalResp.Header.Get("Retry-After") != "":
		return "", fmt.Errorf("too many requests")
	default:
		location := finalResp.Header.Get("Location")
		if location != "" {
			return strings.TrimPrefix(location, redditURLBase), nil
		}
		return "", nil
	}
}

// stream proxies a request to the given URL.
func stream(urlStr string, r *http.Request) (*http.Response, error) {
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't build empty body in stream: %w", err)
	}

	// Copy useful headers
	for _, key := range []string{"Range", "If-Modified-Since", "Cache-Control"} {
		if val := r.Header.Get(key); val != "" {
			req.Header.Set(key, val)
		}
	}

	// Add User-Agent
	d := device.Load().(Device)
	req.Header.Set("User-Agent", d.UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Remove unwanted headers
	for _, key := range []string{
		"access-control-expose-headers", "server", "vary", "etag", "x-cdn",
		"x-cdn-client-region", "x-cdn-name", "x-cdn-server-region",
		"x-reddit-cdn", "x-reddit-video-features", "Nel", "Report-To",
	} {
		resp.Header.Del(key)
	}

	return resp, nil
}

// redditGet makes a GET request to Reddit.
func redditGet(path string, quarantine bool) (*http.Response, error) {
	return request(http.MethodGet, path, true, quarantine, redditURLBase, redditURLBaseHost)
}

// redditShortHead makes a HEAD request to Reddit's short URL.
func redditShortHead(path string, quarantine bool, basePath, host string) (*http.Response, error) {
	return request(http.MethodHead, path, false, quarantine, basePath, host)
}

// request makes a request to Reddit.
func request(method, path string, redirect, quarantine bool, basePath, host string) (*http.Response, error) {
	var fullURL string
	if strings.HasPrefix(path, "http") {
		fullURL = path
	} else {
		fullURL = fmt.Sprintf("%s%s", basePath, path)
	}

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Host":            host,
		"Accept-Encoding": "gzip",
		"Cookie":          "",
	}
	if method != http.MethodGet {
		headers["Accept-Encoding"] = "identity"
	}
	if quarantine {
		headers["Cookie"] = `_options=%7B%22pref_quarantine_optin%22%3A%20true%2C%20%22pref_gated_sr_optin%22%3A%20true%7D`
	}

	// Add OAuth headers
	oc := oauthClient.Load().(Oauth)
	for k, v := range oc.HeadersMap {
		headers[k] = v
	}

	// shuffle headers
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	for _, k := range keys {
		req.Header.Set(k, headers[k])
	}

	// Add User-Agent from device
	d := device.Load().(Device)
	req.Header.Set("User-Agent", d.UserAgent)

	var resp *http.Response
	if redirect {
		// Use default client that follows redirects
		resp, err = http.DefaultClient.Do(req)
	} else {
		// Use our non-redirecting client
		resp, err = httpClient.Do(req)
	}

	if err != nil {
		return nil, err
	}

	// Handle gzip
	if resp.Header.Get("Content-Encoding") == "gzip" {
		body, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = body
	}

	return resp, nil
}

// jsonRequest makes a request to a Reddit API and parses the JSON response.
func jsonRequest(path string, quarantine bool) (interface{}, error) {
	// Rate limit check
	currentRateLimit := atomic.LoadUint32(&oauthRatelimitRemaining)
	if currentRateLimit < 10 && !oauthIsRollingOver.Load() {
		log.Printf("Rate limit %d is low. Spawning force_refresh_token()", currentRateLimit)
		// go forceRefreshToken() // To be implemented
	}
	atomic.AddUint32(&oauthRatelimitRemaining, ^uint32(0)) // Decrement

	resp, err := redditGet(path, quarantine)
	if err != nil {
		return nil, fmt.Errorf("couldn't send request to Reddit: %w | %s", err, path)
	}
	defer resp.Body.Close()

	// Update rate limit info
	if remaining := resp.Header.Get("x-ratelimit-remaining"); remaining != "" {
		if val, err := VParseFloat(remaining); err == nil {
			atomic.StoreUint32(&oauthRatelimitRemaining, uint32(val))
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed receiving body from Reddit: %w | %s", err, path)
	}

	if len(bodyBytes) == 0 {
		// Rate limited
		// go forceRefreshToken()
		if reset := resp.Header.Get("x-ratelimit-reset"); reset != "" {
			return nil, fmt.Errorf("reddit rate limit exceeded. Try refreshing in a few seconds. Rate limit will reset in: %s", reset)
		}
		return nil, fmt.Errorf("reddit rate limit exceeded")
	}

	var data interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("reddit is having issues, check if there's an outage")
		}
		return nil, fmt.Errorf("failed to parse page JSON data: %w | %s", err, path)
	}

	// Check for errors in JSON response, assuming it's a map for error checking
	if dataMap, ok := data.(map[string]interface{}); ok {
		if errVal, ok := dataMap["error"]; ok {
			// Handle different error reasons
			reason, _ := dataMap["reason"].(string)
			message, _ := dataMap["message"].(string)
			switch reason {
			case "quarantined", "gated", "private", "banned":
				return nil, fmt.Errorf("%s", reason)
			}
			if message == "Unauthorized" {
				log.Println("Forcing a token refresh")
				// go forceRefreshToken()
				return nil, fmt.Errorf("oauth token has expired. Please refresh the page")
			}
			return nil, fmt.Errorf("reddit error %v \"%s\": %s | %s", errVal, reason, message, path)
		}
	}


	return data, nil
}

// formatURL is a placeholder for the Rust format_url function
func formatURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL // or handle error appropriately
	}
	// Basic implementation: return path and query
	return u.Path + "?" + u.RawQuery
}

// VParseFloat tries to parse a float from a string, used for header values
func VParseFloat(s string) (float64, error) {
    var f float64
    _, err := fmt.Sscanf(s, "%f", &f)
    return f, err
}