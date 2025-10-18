package main

import (
	"encoding/json"
	"fmt"
)

type Post struct {
	ID            string
	Title         string
	Community     string
	Subreddit     string
	Body          string
	Author        Author
	Permalink     string
	LinkTitle     string
	Poll          *Poll
	Score         string
	ScoreHidden   bool
	UpvoteRatio   int64
	PostType      string
	Flair         Flair
	Flags         Flags
	Thumbnail     Media
	Media         Media
	Domain        string
	RelTime       string
	Created       string
	CreatedTS     int64
	CreatedUTC    int64
	NumDuplicates int
	Comments      string
	NumComments   int
	Gallery       []GalleryMedia
	Awards        Awards
	NSFW          bool
	Spoiler       bool
	Stickied      bool
	OutURL        string
	WSURL         string
}

type Poll struct {
	// Simplified
}

type Media struct {
	URL      string
	AltURL   string
	Width    int
	Height   int
	Poster   string
	DownloadName string
}

type GalleryMedia struct {
	// Simplified
}

type Flags struct {
	Spoiler  bool
	NSFW     bool
	Stickied bool
}

// FetchPosts fetches posts from a given path.
func FetchPosts(path string, quarantine bool) ([]Post, string, error) {
	rawJSON, err := jsonRequest(path, quarantine)
	if err != nil {
		return nil, "", err
	}

	var responseData struct {
		Data struct {
			Children []struct {
				Data json.RawMessage `json:"data"`
			} `json:"children"`
			After string `json:"after"`
		} `json:"data"`
	}

	// The JSON structure is complex, so we unmarshal it into a raw message first
	// and then into a more structured format. This is a common pattern in Go.
	bytes, err := json.Marshal(rawJSON)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal raw json: %w", err)
	}

	if err := json.Unmarshal(bytes, &responseData); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	var posts []Post
	for _, child := range responseData.Data.Children {
		post, err := parsePost(child.Data)
		if err != nil {
			// Log the error but continue processing other posts
			fmt.Printf("failed to parse post: %v\n", err)
			continue
		}
		posts = append(posts, *post)
	}

	return posts, responseData.Data.After, nil
}

func parsePost(data json.RawMessage) (*Post, error) {
	var p Post
	var rawPost map[string]interface{}
	if err := json.Unmarshal(data, &rawPost); err != nil {
		return nil, err
	}

	// This is a simplified version of the Rust `parse_post` function.
	// A full implementation would require mapping all the fields from the JSON.
	p.ID = val(rawPost, "id")
	p.Title = val(rawPost, "title")
	p.Community = val(rawPost, "subreddit")
	p.Permalink = val(rawPost, "permalink")
	p.NSFW, _ = rawPost["over_18"].(bool)
	p.Author.Name = val(rawPost, "author")

	// ... and so on for all the other fields.

	return &p, nil
}

// val needs to be defined here as well, or in a shared utility file.
// For now, let's duplicate it.
func valFromMap(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}