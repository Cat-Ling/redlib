package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SubredditTemplate is the data for the subreddit page.
type SubredditTemplate struct {
	Sub                Subreddit
	Posts              []*Post
	Sort               [2]string
	Ends               [2]string
	Prefs              Preferences
	URL                string
	RedirectURL        string
	IsFiltered         bool
	AllPostsFiltered   bool
	AllPostsHiddenNSFW bool
	NoPosts            bool
}

// WallTemplate is for displaying messages (e.g., for quarantined subs).
type WallTemplate struct {
	Title string
	Sub   string
	Msg   string
	Prefs Preferences
	URL   string
}

func communityHandler(w http.ResponseWriter, r *http.Request) {
	subName, ok := r.Context().Value("sub").(string)
	if !ok {
		subName = "popular"
	}

	sort, ok := r.Context().Value("sort").(string)
	if !ok {
		sort = "hot"
	}

	quarantined := canAccessQuarantine(r, subName)

	sub, err := fetchSubreddit(subName, quarantined)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sub.NSFW && shouldBeNSFWGated(r, r.URL.String()) {
		nsfwLanding(w, r, r.URL.String())
		return
	}

	path := fmt.Sprintf("/r/%s/%s.json?raw_json=1", subName, sort)
	posts, after, err := fetchPosts(path, quarantined)
	if err != nil {
		if err.Error() == "quarantined" || err.Error() == "gated" {
			quarantineWall(w, r, subName, err.Error())
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter posts (placeholder)
	// _, allPostsFiltered := filterPosts(&posts, getFilters(r))

	templateData := SubredditTemplate{
		Sub:   *sub,
		Posts: posts,
		Sort:  [2]string{"hot", ""}, // Placeholder
		Ends:  [2]string{"", after}, // Placeholder
		Prefs: NewPreferences(r),
		URL:   r.URL.String(),
	}

	template(w, templateData)
}

func quarantineWall(w http.ResponseWriter, r *http.Request, sub, restriction string) {
	wall := WallTemplate{
		Title: fmt.Sprintf("r/%s is %s", sub, restriction),
		Msg:   "Please click the button below to continue to this subreddit.",
		URL:   r.URL.String(),
		Sub:   sub,
		Prefs: NewPreferences(r),
	}
	w.WriteHeader(http.StatusForbidden)
	template(w, wall)
}

func addQuarantineExceptionHandler(w http.ResponseWriter, r *http.Request) {
	// Assumes router provides "sub" and "redir"
	sub := ""
	redir := "/"
	cookie := http.Cookie{
		Name:     fmt.Sprintf("allow_quaran_%s", strings.ToLower(sub)),
		Value:    "true",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour), // Session cookie equivalent
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, redir, http.StatusFound)
}

func fetchSubreddit(subName string, quarantined bool) (*Subreddit, error) {
	if subName == "popular" || subName == "all" || strings.Contains(subName, "+") {
		return &Subreddit{Name: subName}, nil
	}

	path := fmt.Sprintf("/r/%s/about.json?raw_json=1", subName)
	resp, err := jsonRequest(path, quarantined)
	if err != nil {
		return nil, err
	}

	data, ok := resp.(map[string]interface{})["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid subreddit data")
	}

	name, _ := data["display_name"].(string)
	desc, _ := data["public_description"].(string)

	return &Subreddit{
		Name:        name,
		Description: desc,
	}, nil
}

func subscriptionsFiltersHandler(w http.ResponseWriter, r *http.Request) {
	// Simplified logic for adding/removing subscriptions/filters
	// Assumes router provides "sub" and action
	sub := ""
	action := ""

	// Get existing subs/filters from cookies
	subsCookie, _ := r.Cookie("subscriptions")
	subs := strings.Split(subsCookie.Value, "+")
	filtersCookie, _ := r.Cookie("filters")
	filters := strings.Split(filtersCookie.Value, "+")

	switch action {
	case "subscribe":
		subs = append(subs, sub)
	case "unsubscribe":
		// remove sub
	case "filter":
		filters = append(filters, sub)
	case "unfilter":
		// remove filter
	}

	// Set new cookies
	http.SetCookie(w, &http.Cookie{Name: "subscriptions", Value: strings.Join(subs, "+")})
	http.SetCookie(w, &http.Cookie{Name: "filters", Value: strings.Join(filters, "+")})

	http.Redirect(w, r, "/r/"+sub, http.StatusFound)
}