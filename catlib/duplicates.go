package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// --- Start of Stubs/Placeholders (will be moved to other files) ---


// Preferences stores user preferences.
type Preferences struct {
	// Placeholder fields
}

func NewPreferences(r *http.Request) Preferences {
	return Preferences{}
}

func errorResponse(w http.ResponseWriter, r *http.Request, msg string) {
	http.Error(w, msg, http.StatusInternalServerError)
}

func canAccessQuarantine(r *http.Request, sub string) bool {
	return true
}

func shouldBeNSFWGated(r *http.Request, url string) bool {
	return false
}

func nsfwLanding(w http.ResponseWriter, r *http.Request, url string) {
	http.Error(w, "NSFW content blocked", http.StatusForbidden)
}

func getFilters(r *http.Request) map[string]struct{} {
	return make(map[string]struct{})
}

func quarantineResponse(w http.ResponseWriter, r *http.Request, sub, reason string) {
	http.Error(w, fmt.Sprintf("Subreddit %s is %s", sub, reason), http.StatusForbidden)
}


// This is a placeholder for a proper template rendering engine.
func template(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Template RENDERED:\n%#v\n", data)
}

// --- End of Stubs/Placeholders ---

// DuplicatesParams contains the parameters in the URL.
type DuplicatesParams struct {
	Before string
	After  string
	Sort   string
}

// DuplicatesTemplate defines the data for rendering the duplicates page.
type DuplicatesTemplate struct {
	Params             DuplicatesParams
	Post               *Post
	Duplicates         []*Post
	Prefs              Preferences
	URL                string
	NumPostsFiltered   uint64
	AllPostsFiltered   bool
}

// duplicatesItem handles requests for duplicate posts.
func duplicatesItem(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	var sub, id string
	// /r/{sub}/duplicates/{id}
	if len(parts) >= 4 && parts[0] == "r" {
		sub = parts[1]
		id = parts[3]
	} else if len(parts) >= 2 && parts[0] == "duplicates" {
		id = parts[1]
	}

	r = r.WithContext(context.WithValue(r.Context(), "sub", sub))
	r = r.WithContext(context.WithValue(r.Context(), "id", id))

	apiPath := fmt.Sprintf("%s.json?%s&raw_json=1", r.URL.Path, r.URL.RawQuery)
	quarantined := canAccessQuarantine(r, sub)

	// This endpoint returns an array, so we need a special request function or handling.
	// For now, let's assume jsonRequest can return a more generic interface{}.
	rawResponse, err := jsonRequest(apiPath, quarantined)
	if err != nil {
		if err.Error() == "quarantined" || err.Error() == "gated" {
			quarantineResponse(w, r, sub, err.Error())
		} else {
			errorResponse(w, r, err.Error())
		}
		return
	}

	responseData, ok := rawResponse.([]interface{})
	if !ok || len(responseData) < 2 {
		errorResponse(w, r, fmt.Sprintf("Invalid JSON response from Reddit, expected array, got %T", rawResponse))
		return
	}

	// First object contains the original post.
	postListing, ok := responseData[0].(map[string]interface{})
	if !ok {
		errorResponse(w, r, "Invalid post listing in response")
		return
	}
	postChildren, ok := postListing["data"].(map[string]interface{})["children"].([]interface{})
	if !ok || len(postChildren) == 0 {
		errorResponse(w, r, "No post found in response")
		return
	}
	post, err := parsePost(postChildren[0].(json.RawMessage))
	if err != nil {
		errorResponse(w, r, fmt.Sprintf("Failed to parse post: %v", err))
		return
	}

	if post.NSFW && shouldBeNSFWGated(r, r.URL.String()) {
		nsfwLanding(w, r, r.URL.String())
		return
	}

	// Second object contains the duplicates.
	duplicatesListing, ok := responseData[1].(map[string]interface{})
	if !ok {
		errorResponse(w, r, "Invalid duplicates listing in response")
		return
	}

	filters := getFilters(r)
	duplicates, numPostsFiltered, allPostsFiltered := parseDuplicates(duplicatesListing, filters)

	// Pagination logic
	var before, after, sort string
	queryParams := r.URL.Query()
	haveBefore := queryParams.Get("before") != ""
	sort = queryParams.Get("sort")

	if len(duplicates) > 0 {
		// This logic is directly ported from the Rust version to handle Reddit API quirks.
		if queryParams.Get("after") != "" {
			before = "t3_" + duplicates[0].ID
		}

		if haveBefore {
			after = "t3_" + duplicates[len(duplicates)-1].ID

			// Re-fetch to check for a previous page.
			newPath := fmt.Sprintf(
				"%s.json?before=t3_%s&sort=%s&limit=1&raw_json=1",
				strings.TrimSuffix(r.URL.Path, "/"),
				duplicates[0].ID,
				url.QueryEscape(sort),
			)
			// The response is an array, so we need to handle it as such.
			rawPrevCheckResponse, err := jsonRequest(newPath, true)
			if err != nil {
				errorResponse(w, r, fmt.Sprintf("Failed to check for previous page: %v", err))
				return
			}

			if prevCheckArr, ok := rawPrevCheckResponse.([]interface{}); ok && len(prevCheckArr) > 1 {
				if prevDuplicatesListing, ok := prevCheckArr[1].(map[string]interface{}); ok {
					if data, ok := prevDuplicatesListing["data"].(map[string]interface{}); ok {
						if children, ok := data["children"].([]interface{}); ok && len(children) > 0 {
							before = "t3_" + duplicates[0].ID
						}
					}
				}
			}
		} else {
			if data, ok := duplicatesListing["data"].(map[string]interface{}); ok {
				if afterVal, ok := data["after"].(string); ok {
					after = afterVal
				}
			}
		}
	}

	templateData := DuplicatesTemplate{
		Params: DuplicatesParams{
			Before: before,
			After:  after,
			Sort:   sort,
		},
		Post:               post,
		Duplicates:         duplicates,
		Prefs:              NewPreferences(r),
		URL:                r.URL.String(),
		NumPostsFiltered:   numPostsFiltered,
		AllPostsFiltered:   allPostsFiltered,
	}

	template(w, templateData)
}

// parseDuplicates parses the list of duplicate posts from the JSON response.
func parseDuplicates(jsonValue map[string]interface{}, filters map[string]struct{}) ([]*Post, uint64, bool) {
	data, ok := jsonValue["data"].(map[string]interface{})
	if !ok {
		return []*Post{}, 0, false
	}

	children, ok := data["children"].([]interface{})
	if !ok {
		return []*Post{}, 0, false
	}

	var duplicates []*Post
	for _, child := range children {
		post, err := parsePost(child.(json.RawMessage))
		if err == nil {
			duplicates = append(duplicates, post)
		}
	}

	// Placeholder for filterPosts
	// numPostsFiltered, allPostsFiltered := filterPosts(&duplicates, filters)
	return duplicates, 0, false
}