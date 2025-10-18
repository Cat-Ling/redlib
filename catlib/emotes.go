package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	redditEmoteIDNumberRegex = regexp.MustCompile(`"emote\|.*\|(.*)"`)
)

func rewriteEmotes(mediaMetadata map[string]interface{}, commentHTML string) string {
	if mediaMetadata == nil {
		return commentHTML
	}

	var toReplace []struct {
		id   string
		link string
		size string
	}

	for _, data := range mediaMetadata {
		emoteData, ok := data.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := emoteData["id"].(string)
		s, _ := emoteData["s"].(map[string]interface{})
		u, _ := s["u"].(string)
		y, _ := s["y"].(float64)

		if id != "" && u != "" {
			if matches := redditEmoteIDNumberRegex.FindStringSubmatch(id); len(matches) > 1 {
				toReplace = append(toReplace, struct {
					id   string
					link string
					size string
				}{
					id:   fmt.Sprintf(":%s:", matches[1]),
					link: formatRawURL(u),
					size: fmt.Sprintf("%.0f", y),
				})
			}
		}
	}

	result := commentHTML
	for _, replacement := range toReplace {
		imgTag := fmt.Sprintf(`<img loading="lazy" src="%s" width="%s" height="%s" style="vertical-align:text-bottom">`, replacement.link, replacement.size, replacement.size)
		result = strings.ReplaceAll(result, replacement.id, imgTag)
	}

	return result
}