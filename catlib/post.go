package main

import (
	"context"
	"fmt"
	"html"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"
	"time"
)

// --- Start of Stubs/Placeholders ---

// Comment represents a single comment.
type Comment struct {
	ID           string
	Body         string
	Author       Author
	Score        string
	ScoreHidden  bool
	RelTime      string
	Created      string
	Edited       string
	Replies      []Comment
	Highlighted  bool
	Awards       Awards
	Collapsed    bool
	IsFiltered   bool
	MoreCount    int64
	Prefs        Preferences
	PostLink     string
	PostAuthor   string
	ParentID     string
	ParentKind   string
}

// Author represents a Reddit user.
type Author struct {
	Name          string
	Flair         Flair
	Distinguished string
}

// Flair represents user or link flair.
type Flair struct {
	Text            string
	BackgroundColor string
	ForegroundColor string
}

// Awards represents awards on a post or comment.
type Awards struct {
	// Placeholder
}

// --- End of Stubs/Placeholders ---

// PostTemplate is the data for the post page.
type PostTemplate struct {
	Comments         []Comment
	Post             *Post
	Sort             string
	Prefs            Preferences
	SingleThread     bool
	URL              string
	URLWithoutQuery  string
	CommentQuery     string
}

var commentSearchCapture = regexp.MustCompile(`\?q=(.*)&type=comment`)

func postItem(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	var sub, id, commentID string

	// This is a very basic router, a real one would be better
	if parts[0] == "r" && len(parts) > 3 {
		sub = parts[1]
		id = parts[3]
		if len(parts) > 5 {
			commentID = parts[5]
		}
	} else if parts[0] == "comments" && len(parts) > 1 {
		id = parts[1]
		if len(parts) > 3 {
			commentID = parts[3]
		}
	}

	r = r.WithContext(context.WithValue(r.Context(), "sub", sub))
	r = r.WithContext(context.WithValue(r.Context(), "id", id))
	r = r.WithContext(context.WithValue(r.Context(), "comment_id", commentID))

	path := fmt.Sprintf("%s.json?%s&raw_json=1", r.URL.Path, r.URL.RawQuery)
	quarantined := canAccessQuarantine(r, sub)
	url := r.URL.String()

	sort := r.URL.Query().Get("sort")
	singleThread := commentID != ""

	rawResponse, err := jsonRequest(path, quarantined)
	if err != nil {
		// Error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, ok := rawResponse.([]interface{})
	if !ok || len(response) < 2 {
		http.Error(w, "Invalid API response", http.StatusInternalServerError)
		return
	}

	postData, ok := response[0].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid post data", http.StatusInternalServerError)
		return
	}
	post, err := parsePost(postData["data"].(map[string]interface{})["children"].([]interface{})[0])
	if err != nil {
		http.Error(w, "Failed to parse post", http.StatusInternalServerError)
		return
	}

	if post.NSFW && shouldBeNSFWGated(r, url) {
		nsfwLanding(w, r, url)
		return
	}

	commentData, ok := response[1].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid comment data", http.StatusInternalServerError)
		return
	}

	var comments []Comment
	query := ""
	if matches := commentSearchCapture.FindStringSubmatch(url); len(matches) > 1 {
		query, _ = neturl.QueryUnescape(matches[1])
		comments = queryComments(commentData, post.ID, post.Author.Name, commentID, getFilters(r), query, r)
	} else {
		comments = parseComments(commentData, post.ID, post.Author.Name, commentID, getFilters(r), r)
	}

	templateData := PostTemplate{
		Comments:         comments,
		Post:             post,
		Sort:             sort,
		Prefs:            NewPreferences(r),
		SingleThread:     singleThread,
		URL:              url,
		URLWithoutQuery:  strings.Split(url, "?")[0],
		CommentQuery:     query,
	}

	// Placeholder for template rendering
	template(w, templateData)
}

func parseComments(jsonValue map[string]interface{}, postLink, postAuthor, highlightedComment string, filters map[string]struct{}, r *http.Request) []Comment {
	var comments []Comment
	children, ok := jsonValue["data"].(map[string]interface{})["children"].([]interface{})
	if !ok {
		return comments
	}

	for _, c := range children {
		commentMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		data, ok := commentMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		var replies []Comment
		if repliesData, ok := data["replies"].(map[string]interface{}); ok {
			replies = parseComments(repliesData, postLink, postAuthor, highlightedComment, filters, r)
		}
		comments = append(comments, buildComment(commentMap, data, replies, postLink, postAuthor, highlightedComment, filters, r))
	}
	return comments
}

func queryComments(jsonValue map[string]interface{}, postLink, postAuthor, highlightedComment string, filters map[string]struct{}, query string, r *http.Request) []Comment {
	// Simplified version
	var results []Comment
	allComments := parseComments(jsonValue, postLink, postAuthor, highlightedComment, filters, r)
	for _, c := range allComments {
		if strings.Contains(strings.ToLower(c.Body), strings.ToLower(query)) {
			results = append(results, c)
		}
	}
	return results
}

func buildComment(comment, data map[string]interface{}, replies []Comment, postLink, postAuthor, highlightedComment string, filters map[string]struct{}, r *http.Request) Comment {
	id, _ := data["id"].(string)
	authorName, _ := data["author"].(string)
	body, _ := data["body_html"].(string)

	body = html.UnescapeString(body) // The rust version does more complex rewriting

	createdUTC, _ := data["created_utc"].(float64)
	relTime, created := formatTime(createdUTC)

	score, _ := data["score"].(float64)
	scoreHidden, _ := data["score_hidden"].(bool)

	return Comment{
		ID:          id,
		Body:        body,
		Author:      Author{Name: authorName},
		Score:       fmt.Sprintf("%.0f", score),
		ScoreHidden: scoreHidden,
		RelTime:     relTime,
		Created:     created,
		Replies:     replies,
		Highlighted: id == highlightedComment,
		// ... other fields
	}
}

func formatTime(utc float64) (string, string) {
	t := time.Unix(int64(utc), 0)
	return time.Since(t).String(), t.Format(time.RFC1123)
}