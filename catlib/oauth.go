package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	redditAndroidOauthClientID = "ohXpoqrZYub1kg"
	authEndpoint               = "https://www.reddit.com"
	oauthTimeout               = 5 * time.Second
)

// Oauth handles OAuth2 authentication with Reddit.
type Oauth struct {
	InitialHeaders map[string]string
	HeadersMap     map[string]string
	Token          string
	ExpiresIn      int64
	device         Device
}

// Device represents a spoofed client device.
type Device struct {
	OauthID       string
	InitialHeaders map[string]string
	Headers       map[string]string
	UserAgent     string
}

// newOauth creates a new OAuth client.
func newOauth(device Device) *Oauth {
	for {
		oauth, err := newOauthWithTimeout(device)
		if err != nil {
			log.Printf("Failed to create OAuth client: %v. Retrying in 5 seconds...", err)
			time.Sleep(oauthTimeout)
			continue
		}
		log.Println("[✅] Successfully created OAuth client")
		return oauth
	}
}

func newOauthWithTimeout(device Device) (*Oauth, error) {
	oauth := &Oauth{
		device:         device,
		InitialHeaders: device.InitialHeaders,
		HeadersMap:     device.Headers,
	}

	c := make(chan error, 1)
	go func() {
		c <- oauth.login()
	}()

	select {
	case err := <-c:
		if err != nil {
			return nil, err
		}
		return oauth, nil
	case <-time.After(oauthTimeout):
		return nil, fmt.Errorf("oauth login timed out")
	}
}

func (o *Oauth) login() error {
	url := fmt.Sprintf("%s/auth/v2/oauth/access-token/loid", authEndpoint)
	payload := map[string][]string{"scopes": {"*", "email", "pii"}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))

	for key, value := range o.device.InitialHeaders {
		req.Header.Set(key, value)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(o.device.OauthID + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if loid := resp.Header.Get("x-reddit-loid"); loid != "" {
		o.HeadersMap["x-reddit-loid"] = loid
	}
	if session := resp.Header.Get("x-reddit-session"); session != "" {
		o.HeadersMap["x-reddit-session"] = session
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return fmt.Errorf("access_token not found in response")
	}
	o.Token = accessToken

	expiresIn, ok := result["expires_in"].(float64)
	if !ok {
		return fmt.Errorf("expires_in not found in response")
	}
	o.ExpiresIn = int64(expiresIn)

	o.HeadersMap["Authorization"] = "Bearer " + o.Token
	log.Printf("[✅] Success - Retrieved token, expires in %d", o.ExpiresIn)
	return nil
}

// tokenDaemon runs in the background to refresh the OAuth token.
func tokenDaemon() {
	for {
		expiresIn := oauthClient.Load().(*Oauth).ExpiresIn
		duration := time.Duration(expiresIn-120) * time.Second
		log.Printf("[⏳] Waiting for %v seconds before refreshing OAuth token...", duration)
		time.Sleep(duration)
		log.Printf("[⌛] %v Elapsed! Refreshing OAuth token...", duration)
		forceRefreshToken()
	}
}

func forceRefreshToken() {
	if !oauthIsRollingOver.CompareAndSwap(false, true) {
		log.Println("Skipping refresh token roll over, already in progress")
		return
	}
	defer oauthIsRollingOver.Store(false)

	log.Printf("Rolling over refresh token. Current rate limit: %d", atomic.LoadUint32(&oauthRatelimitRemaining))

	newDevice := newDevice()
	device.Store(newDevice)

	newClient := newOauth(newDevice)
	oauthClient.Store(newClient)
	atomic.StoreUint32(&oauthRatelimitRemaining, 99)
}

// newDevice creates a new spoofed Android device.
func newDevice() Device {
	id := uuid.New().String()
	androidAppVersion := androidAppVersionList[rand.Intn(len(androidAppVersionList))]
	androidVersion := rand.Intn(6) + 9 // 9-14
	userAgent := fmt.Sprintf("Reddit/%s/Android %d", androidAppVersion, androidVersion)
	qos := float32(rand.Intn(99001)+1000) / 1000.0

	headers := map[string]string{
		"User-Agent":           userAgent,
		"x-reddit-retry":       "algo=no-retries",
		"x-reddit-compression": "1",
		"x-reddit-qos":         fmt.Sprintf("%.3f", qos),
		"Content-Type":         "application/json; charset=UTF-8",
		"client-vendor-id":     id,
		"X-Reddit-Device-Id":   id,
	}

	log.Printf("[🔄] Spoofing Android client with User-Agent: %s", userAgent)

	return Device{
		OauthID:       redditAndroidOauthClientID,
		Headers:       headers,
		InitialHeaders: headers,
		UserAgent:     userAgent,
	}
}


var (
	oauthClientOnce sync.Once
)

func initializeOAuth() {
	oauthClientOnce.Do(func() {
		d := newDevice()
		device.Store(d)
		oc := newOauth(d)
		oauthClient.Store(oc)
		go tokenDaemon()
	})
}