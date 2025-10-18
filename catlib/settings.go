package main

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SettingsTemplate is the data for the settings page.
type SettingsTemplate struct {
	Prefs Preferences
	URL   string
}

var prefsKeys = []string{
	"theme", "front_page", "layout", "wide", "comment_sort", "post_sort",
	"blur_spoiler", "show_nsfw", "blur_nsfw", "use_hls", "hide_hls_notification",
	"autoplay_videos", "hide_sidebar_and_summary", "fixed_navbar", "hide_awards",
	"hide_score", "disable_visit_reddit_confirmation", "video_quality", "remove_default_feeds",
}

// settingsGetHandler displays the settings page.
func settingsGetHandler(w http.ResponseWriter, r *http.Request) {
	templateData := SettingsTemplate{
		Prefs: NewPreferences(r),
		URL:   r.URL.String(),
	}
	template(w, templateData)
}

// settingsSetHandler updates the settings cookies.
func settingsSetHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	form, _ := url.ParseQuery(string(body))

	for _, name := range prefsKeys {
		if value, ok := form[name]; ok {
			cookie := http.Cookie{
				Name:     name,
				Value:    value[0],
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(365 * 24 * time.Hour),
			}
			http.SetCookie(w, &cookie)
		} else {
			// Remove cookie
			cookie := http.Cookie{
				Name:   name,
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			}
			http.SetCookie(w, &cookie)
		}
	}

	// Simplified subscription/filter handling
	handleChunkedCookie(w, r, form, "subscriptions")
	handleChunkedCookie(w, r, form, "filters")

	http.Redirect(w, r, "/settings", http.StatusFound)
}

// Simplified version of the chunked cookie logic
func handleChunkedCookie(w http.ResponseWriter, r *http.Request, form url.Values, key string) {
	if values, ok := form[key]; ok {
		// In a real implementation, we'd split this into multiple cookies if it's too large.
		cookie := http.Cookie{
			Name:     key,
			Value:    values[0],
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(365 * 24 * time.Hour),
		}
		http.SetCookie(w, &cookie)

		// Naive removal of old numbered cookies
		for i := 1; i < 10; i++ {
			http.SetCookie(w, &http.Cookie{Name: key + strconv.Itoa(i), MaxAge: -1})
		}
	} else {
		// Remove all cookies for this key
		http.SetCookie(w, &http.Cookie{Name: key, MaxAge: -1})
		for i := 1; i < 10; i++ {
			http.SetCookie(w, &http.Cookie{Name: key + strconv.Itoa(i), MaxAge: -1})
		}
	}
}

func settingsRestoreHandler(w http.ResponseWriter, r *http.Request) {
	// Logic from set_cookies_method with remove_cookies=true
	query := r.URL.Query()
	redirectPath := "/"
	if redirect, ok := query["redirect"]; ok && len(redirect) > 0 {
		redirectPath = redirect[0]
		if !strings.HasPrefix(redirectPath, "/") {
			redirectPath = "/" + redirectPath
		}
	}

	for _, name := range prefsKeys {
		if value, ok := query[name]; ok {
			cookie := http.Cookie{
				Name:    name,
				Value:   value[0],
				Path:    "/",
				Expires: time.Now().Add(365 * 24 * time.Hour),
			}
			http.SetCookie(w, &cookie)
		} else {
			// Remove cookie
			cookie := http.Cookie{Name: name, MaxAge: -1}
			http.SetCookie(w, &cookie)
		}
	}
	// Simplified subscription/filter handling
	handleChunkedCookie(w, r, query, "subscriptions")
	handleChunkedCookie(w, r, query, "filters")

	http.Redirect(w, r, redirectPath, http.StatusFound)
}