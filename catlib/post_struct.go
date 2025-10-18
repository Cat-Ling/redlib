package main

// Post represents a single Reddit post.
type Post struct {
	ID          string
	Title       string
	Subreddit   string
	Author      Author
	Permalink   string
	Body        string
	Score       string
	UpvoteRatio int64
	PostType    string
	Flair       Flair
	NSFW        bool
	Stickied    bool
	Spoiler     bool
	Awards      Awards
	Media       Media
	Gallery     []Media
	Created     string
	CreatedUTC  int64
}

// Media represents media associated with a post.
type Media struct {
	URL      string
	AltURL   string
	Width    int64
	Height   int64
	Poster   string
	DownloadName string
}