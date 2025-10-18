package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// --- Struct definitions moved from other files ---

// --- End of Struct definitions ---

// --- Utility functions ---

func redirect(w http.ResponseWriter, r *http.Request, path string) {
	http.Redirect(w, r, path, http.StatusFound)
}

// val gets a string value from a nested map.
func val(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	if d, ok := data["data"].(map[string]interface{}); ok {
		if v, ok := d[key].(string); ok {
			return v
		}
	}
	return ""
}

// format_num formats large numbers.
func formatNum(num int64) (string, string) {
	full := fmt.Sprintf("%d", num)
	if num >= 1_000_000 || num <= -1_000_000 {
		return fmt.Sprintf("%.1fm", float64(num)/1_000_000.0), full
	}
	if num >= 1000 || num <= -1000 {
		return fmt.Sprintf("%.1fk", float64(num)/1000.0), full
	}
	return full, full
}

// rewriteUrls rewrites Reddit URLs to local URLs in HTML content.
func rewriteUrls(inputText string) string {
	// These regexes are for URLs inside href attributes, etc.
	redditRegex := regexp.MustCompile(`href="(https?://)(www\.|old\.|np\.|amp\.|new\.|)(reddit\.com|redd\.it)/`)
	text1 := redditRegex.ReplaceAllString(inputText, `href="/`)

	// Remove backslashes from URLs
	text1 = strings.ReplaceAll(text1, "%5C", "")
	text1 = strings.ReplaceAll(text1, "\\_", "_")

	// The logic for rewriting preview links and emotes is complex and will be implemented next.
	// For now, this handles basic reddit links.
	return text1
}

// Regular expressions for format_url
var (
	regexURLWww           = regexp.MustCompile(`https?://www\.reddit\.com/(.*)`)
	regexURLOld           = regexp.MustCompile(`https?://old\.reddit\.com/(.*)`)
	regexURLNp            = regexp.MustCompile(`https?://np\.reddit\.com/(.*)`)
	regexURLPlain         = regexp.MustCompile(`https?://reddit\.com/(.*)`)
	regexURLVideos        = regexp.MustCompile(`https?://v\.redd\.it/(.*)/DASH_([0-9]{2,4}(\.mp4|$|\?source=fallback))`)
	regexURLVideosHLS     = regexp.MustCompile(`https?://v\.redd\.it/(.+)/(HLSPlaylist\.m3u8.*)$`)
	regexURLImages        = regexp.MustCompile(`https?://i\.redd\.it/(.*)`)
	regexURLThumbsA       = regexp.MustCompile(`https?://a\.thumbs\.redditmedia\.com/(.*)`)
	regexURLThumbsB       = regexp.MustCompile(`https?://b\.thumbs\.redditmedia\.com/(.*)`)
	regexURLEmoji         = regexp.MustCompile(`https?://emoji\.redditmedia\.com/(.*)/(.*)`)
	regexURLPreview       = regexp.MustCompile(`https?://preview\.redd\.it/(.*)`)
	regexURLExternalPreview = regexp.MustCompile(`https?://external-preview\.redd\.it/(.*)`)
	regexURLStyles        = regexp.MustCompile(`https?://styles\.redditmedia\.com/(.*)`)
	regexURLStaticMedia   = regexp.MustCompile(`https?://www\.redditstatic\.com/(.*)`)
)

func formatRawURL(rawURL string) string {
	if rawURL == "" || rawURL == "self" || rawURL == "default" || rawURL == "nsfw" || rawURL == "spoiler" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	domain := u.Hostname()

	capture := func(re *regexp.Regexp, format string, segments int) string {
		matches := re.FindStringSubmatch(rawURL)
		if len(matches) > segments {
			if segments == 1 {
				return fmt.Sprintf(format, matches[1])
			}
			if segments == 2 {
				return fmt.Sprintf(format, matches[1], matches[2])
			}
		}
		return ""
	}

	switch domain {
	case "www.reddit.com":
		return capture(regexURLWww, "/%s", 1)
	case "old.reddit.com":
		return capture(regexURLOld, "/%s", 1)
	case "np.reddit.com":
		return capture(regexURLNp, "/%s", 1)
	case "reddit.com":
		return capture(regexURLPlain, "/%s", 1)
	case "v.redd.it":
		if result := capture(regexURLVideos, "/vid/%s/%s", 2); result != "" {
			return result
		}
		return capture(regexURLVideosHLS, "/hls/%s/%s", 2)
	case "i.redd.it":
		return capture(regexURLImages, "/img/%s", 1)
	case "a.thumbs.redditmedia.com":
		return capture(regexURLThumbsA, "/thumb/a/%s", 1)
	case "b.thumbs.redditmedia.com":
		return capture(regexURLThumbsB, "/thumb/b/%s", 1)
	case "emoji.redditmedia.com":
		return capture(regexURLEmoji, "/emoji/%s/%s", 2)
	case "preview.redd.it":
		return capture(regexURLPreview, "/preview/pre/%s", 1)
	case "external-preview.redd.it":
		return capture(regexURLExternalPreview, "/preview/external-pre/%s", 1)
	case "styles.redditmedia.com":
		return capture(regexURLStyles, "/style/%s", 1)
	case "www.redditstatic.com":
		return capture(regexURLStaticMedia, "/static/%s", 1)
	default:
		return rawURL
	}
}

// setting gets a setting from cookies or returns a default.
func getCookieSetting(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		// Return default from config if available
		if val := GetSetting("REDLIB_DEFAULT_" + strings.ToUpper(name)); val != nil {
			return *val
		}
		return ""
	}
	return cookie.Value
}
func getFullURL() string {
	if val := GetSetting("REDLIB_FULL_URL"); val != nil {
		return *val
	}
	return ""
}
func getPostURL(post *Post) string {
	// Simplified version
	return getFullURL() + post.ID
}