package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// --- Start of Stubs/Placeholders ---

func getSetting(r *http.Request, key string) string {
	// Placeholder
	return ""
}

func sfwOnly() bool {
	// Placeholder
	return false
}

// --- End of Stubs/Placeholders ---

// SearchParams holds the search query parameters.
type SearchParams struct {
	Q          string
	Sort       string
	T          string
	Before     string
	After      string
	RestrictSR string
	Typed      string
}

// Subreddit holds information about a subreddit search result.
type Subreddit struct {
	Name        string
	URL         string
	Icon        string
	Description string
	Subscribers string
	NSFW        bool
}

// SearchTemplate is the data for the search results page.
type SearchTemplate struct {
	Posts               []*Post
	Subreddits          []Subreddit
	Sub                 string
	Params              SearchParams
	Prefs               Preferences
	URL                 string
	IsFiltered          bool
	AllPostsFiltered    bool
	AllPostsHiddenNSFW  bool
	NoPosts             bool
}

var redditURLMatch = regexp.MustCompile(`^https?://([^\./]+\.)*reddit.com/`)

func findHandler(w http.ResponseWriter, r *http.Request) {
	query, _ := url.ParseQuery(r.URL.RawQuery)
	q := query.Get("q")
	r = r.WithContext(context.WithValue(r.Context(), "q", q))

	nsfwResults := ""
	if getSetting(r, "show_nsfw") == "on" && !sfwOnly() {
		nsfwResults = "&include_over_18=on"
	}

	uriPath := strings.Replace(r.URL.Path, "+", "%2B", -1)
	path := fmt.Sprintf("%s.json?%s%s&raw_json=1", uriPath, r.URL.RawQuery, nsfwResults)

	q = redditURLMatch.ReplaceAllString(q, "")

	if q == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if strings.HasPrefix(q, "r/") || strings.HasPrefix(q, "user/") {
		http.Redirect(w, r, "/"+q, http.StatusFound)
		return
	}

	// Simplified logic for subreddits
	// Assumes router param "sub"
	sub := ""
	quarantined := canAccessQuarantine(r, sub)

	typed := query.Get("type")
	sort := query.Get("sort")
	if sort == "" {
		sort = "relevance"
	}
	// filters := getFilters(r)

	var subreddits []Subreddit
	if query.Get("restrict_sr") == "" {
		subreddits = searchSubreddits(q, typed)
		// Filtering logic would go here
	}

	// Placeholder for fetching posts
	// For now, we'll just render the template with empty posts
	posts, after, err := fetchPosts(path, quarantined)
	if err != nil {
		// handle error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateData := SearchTemplate{
		Posts:      posts,
		Subreddits: subreddits,
		Sub:        sub,
		Params: SearchParams{
			Q:          q,
			Sort:       sort,
			T:          query.Get("t"),
			RestrictSR: query.Get("restrict_sr"),
			Typed:      typed,
			After:      after,
		},
		Prefs: NewPreferences(r),
		URL:   r.URL.String(),
	}

	template(w, templateData)
}

func searchSubreddits(q, typed string) []Subreddit {
	limit := "3"
	if typed == "sr_user" {
		limit = "50"
	}
	path := fmt.Sprintf("/subreddits/search.json?q=%s&limit=%s", url.QueryEscape(q), limit)

	rawResponse, err := jsonRequest(path, false)
	if err != nil {
		return []Subreddit{}
	}

	response, ok := rawResponse.(map[string]interface{})
	if !ok {
		return []Subreddit{}
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return []Subreddit{}
	}

	children, ok := data["children"].([]interface{})
	if !ok {
		return []Subreddit{}
	}

	var subreddits []Subreddit
	for _, child := range children {
		srData, ok := child.(map[string]interface{})["data"].(map[string]interface{})
		if !ok {
			continue
		}

		icon, _ := srData["icon_img"].(string)
		if icon == "" {
			icon, _ = srData["community_icon"].(string)
		}

		name, _ := srData["display_name"].(string)
		desc, _ := srData["public_description"].(string)
		subs, _ := srData["subscribers"].(float64)

		subreddits = append(subreddits, Subreddit{
			Name:        name,
			URL:         fmt.Sprintf("/r/%s", name),
			Icon:        icon,
			Description: desc,
			Subscribers: fmt.Sprintf("%.0f", subs),
		})
	}

	return subreddits
}

// Placeholder for Post.fetch
func fetchPosts(path string, quarantined bool) ([]*Post, string, error) {
	return []*Post{}, "", nil
}