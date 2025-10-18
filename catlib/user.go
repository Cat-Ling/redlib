package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// User holds information about a Reddit user.
type User struct {
	Name        string
	Title       string
	Icon        string
	Karma       int64
	Created     string
	Banner      string
	Description string
	NSFW        bool
}

// UserTemplate is the data for the user profile page.
type UserTemplate struct {
	User               User
	Posts              []*Post
	Sort               [2]string
	Ends               [2]string
	Listing            string
	Prefs              Preferences
	URL                string
	RedirectURL        string
	IsFiltered         bool
	AllPostsFiltered   bool
	AllPostsHiddenNSFW bool
	NoPosts            bool
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/user/"), "/")
	name := parts[0]
	listing := "overview"
	if len(parts) > 1 {
		listing = parts[1]
	}
	r = r.WithContext(context.WithValue(r.Context(), "name", name))
	r = r.WithContext(context.WithValue(r.Context(), "listing", listing))

	path := fmt.Sprintf("/user/%s/%s.json?%s&raw_json=1", name, listing, r.URL.RawQuery)

	user, err := fetchUser(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.NSFW && shouldBeNSFWGated(r, r.URL.String()) {
		nsfwLanding(w, r, r.URL.String())
		return
	}

	posts, after, err := fetchPosts(path, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filtering logic would go here

	templateData := UserTemplate{
		User:    *user,
		Posts:   posts,
		Listing: listing,
		Sort:    [2]string{r.URL.Query().Get("sort"), r.URL.Query().Get("t")},
		Ends:    [2]string{r.URL.Query().Get("after"), after},
		Prefs:   NewPreferences(r),
		URL:     r.URL.String(),
	}

	template(w, templateData)
}

func fetchUser(name string) (*User, error) {
	path := fmt.Sprintf("/user/%s/about.json?raw_json=1", name)
	resp, err := jsonRequest(path, false)
	if err != nil {
		return nil, err
	}

	data, ok := resp.(map[string]interface{})["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid user data")
	}

	createdUTC, _ := data["created"].(float64)
	created := time.Unix(int64(createdUTC), 0)

	subredditData, _ := data["subreddit"].(map[string]interface{})
	title, _ := subredditData["title"].(string)
	icon, _ := subredditData["icon_img"].(string)
	banner, _ := subredditData["banner_img"].(string)
	desc, _ := subredditData["public_description"].(string)
	nsfw, _ := subredditData["over_18"].(bool)
	karma, _ := data["total_karma"].(float64)

	return &User{
		Name:        name,
		Title:       title,
		Icon:        icon,
		Karma:       int64(karma),
		Created:     created.Format("Jan 02 '06"),
		Banner:      banner,
		Description: desc,
		NSFW:        nsfw,
	}, nil
}