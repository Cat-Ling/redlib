package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

const (
	DefaultPushshiftFrontend = "undelete.pullpush.io"
)

// Config stores the configuration parsed from the environment variables and the
// config file.
type Config struct {
	SFWOnly                            *string `toml:"REDLIB_SFW_ONLY" json:"sfw_only,omitempty"`
	DefaultTheme                       *string `toml:"REDLIB_DEFAULT_THEME" json:"default_theme,omitempty"`
	DefaultFrontPage                   *string `toml:"REDLIB_DEFAULT_FRONT_PAGE" json:"default_front_page,omitempty"`
	DefaultLayout                      *string `toml:"REDLIB_DEFAULT_LAYOUT" json:"default_layout,omitempty"`
	DefaultWide                        *string `toml:"REDLIB_DEFAULT_WIDE" json:"default_wide,omitempty"`
	DefaultCommentSort                 *string `toml:"REDLIB_DEFAULT_COMMENT_SORT" json:"default_comment_sort,omitempty"`
	DefaultPostSort                    *string `toml:"REDLIB_DEFAULT_POST_SORT" json:"default_post_sort,omitempty"`
	DefaultBlurSpoiler                 *string `toml:"REDLIB_DEFAULT_BLUR_SPOILER" json:"default_blur_spoiler,omitempty"`
	DefaultShowNSFW                    *string `toml:"REDLIB_DEFAULT_SHOW_NSFW" json:"default_show_nsfw,omitempty"`
	DefaultBlurNSFW                    *string `toml:"REDLIB_DEFAULT_BLUR_NSFW" json:"default_blur_nsfw,omitempty"`
	DefaultUseHLS                      *string `toml:"REDLIB_DEFAULT_USE_HLS" json:"default_use_hls,omitempty"`
	DefaultHideHLSNotification         *string `toml:"REDLIB_DEFAULT_HIDE_HLS_NOTIFICATION" json:"default_hide_hls_notification,omitempty"`
	DefaultHideAwards                  *string `toml:"REDLIB_DEFAULT_HIDE_AWARDS" json:"default_hide_awards,omitempty"`
	DefaultHideSidebarAndSummary       *string `toml:"REDLIB_DEFAULT_HIDE_SIDEBAR_AND_SUMMARY" json:"default_hide_sidebar_and_summary,omitempty"`
	DefaultHideScore                   *string `toml:"REDLIB_DEFAULT_HIDE_SCORE" json:"default_hide_score,omitempty"`
	DefaultSubscriptions               *string `toml:"REDLIB_DEFAULT_SUBSCRIPTIONS" json:"default_subscriptions,omitempty"`
	DefaultFilters                     *string `toml:"REDLIB_DEFAULT_FILTERS" json:"default_filters,omitempty"`
	DefaultDisableVisitRedditConfirmation *string `toml:"REDLIB_DEFAULT_DISABLE_VISIT_REDDIT_CONFIRMATION" json:"default_disable_visit_reddit_confirmation,omitempty"`
	Banner                             *string `toml:"REDLIB_BANNER" json:"banner,omitempty"`
	RobotsDisableIndexing              *string `toml:"REDLIB_ROBOTS_DISABLE_INDEXING" json:"robots_disable_indexing,omitempty"`
	Pushshift                          *string `toml:"REDLIB_PUSHSHIFT_FRONTEND" json:"pushshift,omitempty"`
	EnableRSS                          *string `toml:"REDLIB_ENABLE_RSS" json:"enable_rss,omitempty"`
	FullURL                            *string `toml:"REDLIB_FULL_URL" json:"full_url,omitempty"`
	DefaultRemoveDefaultFeeds          *string `toml:"REDLIB_DEFAULT_REMOVE_DEFAULT_FEEDS" json:"default_remove_default_feeds,omitempty"`
}


var (
	globalConfig *Config
	configOnce   sync.Once
)

// loadConfigFromFile loads the configuration from a file.
func loadConfigFromFile(name string) (*Config, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var config Config
	if _, err := toml.Decode(string(data), &config); err != nil {
		log.Printf("could not decode toml config %s: %v", name, err)
		return nil, err
	}
	return &config, nil
}

// LoadGlobalConfig loads the configuration from environment variables and config files.
func LoadGlobalConfig() *Config {
	configOnce.Do(func() {
		fileConfig, err := loadConfigFromFile("redlib.toml")
		if err != nil {
			fileConfig, _ = loadConfigFromFile("libreddit.toml")
		}

		parse := func(key string, fileValue *string) *string {
			// Env var takes precedence
			if value, ok := os.LookupEnv(key); ok {
				return &value
			}
			// Legacy env var
			legacyKey := strings.Replace(key, "REDLIB_", "LIBREDDIT_", 1)
			if value, ok := os.LookupEnv(legacyKey); ok {
				return &value
			}
			// Then file value
			return fileValue
		}

		c := &Config{}
		if fileConfig != nil {
			c = fileConfig
		}

		globalConfig = &Config{
			SFWOnly:                            parse("REDLIB_SFW_ONLY", c.SFWOnly),
			DefaultTheme:                       parse("REDLIB_DEFAULT_THEME", c.DefaultTheme),
			DefaultFrontPage:                   parse("REDLIB_DEFAULT_FRONT_PAGE", c.DefaultFrontPage),
			DefaultLayout:                      parse("REDLIB_DEFAULT_LAYOUT", c.DefaultLayout),
			DefaultPostSort:                    parse("REDLIB_DEFAULT_POST_SORT", c.DefaultPostSort),
			DefaultWide:                        parse("REDLIB_DEFAULT_WIDE", c.DefaultWide),
			DefaultCommentSort:                 parse("REDLIB_DEFAULT_COMMENT_SORT", c.DefaultCommentSort),
			DefaultBlurSpoiler:                 parse("REDLIB_DEFAULT_BLUR_SPOILER", c.DefaultBlurSpoiler),
			DefaultShowNSFW:                    parse("REDLIB_DEFAULT_SHOW_NSFW", c.DefaultShowNSFW),
			DefaultBlurNSFW:                    parse("REDLIB_DEFAULT_BLUR_NSFW", c.DefaultBlurNSFW),
			DefaultUseHLS:                      parse("REDLIB_DEFAULT_USE_HLS", c.DefaultUseHLS),
			DefaultHideHLSNotification:         parse("REDLIB_DEFAULT_HIDE_HLS_NOTIFICATION", c.DefaultHideHLSNotification),
			DefaultHideAwards:                  parse("REDLIB_DEFAULT_HIDE_AWARDS", c.DefaultHideAwards),
			DefaultHideSidebarAndSummary:       parse("REDLIB_DEFAULT_HIDE_SIDEBAR_AND_SUMMARY", c.DefaultHideSidebarAndSummary),
			DefaultHideScore:                   parse("REDLIB_DEFAULT_HIDE_SCORE", c.DefaultHideScore),
			DefaultSubscriptions:               parse("REDLIB_DEFAULT_SUBSCRIPTIONS", c.DefaultSubscriptions),
			DefaultFilters:                     parse("REDLIB_DEFAULT_FILTERS", c.DefaultFilters),
			DefaultDisableVisitRedditConfirmation: parse("REDLIB_DEFAULT_DISABLE_VISIT_REDDIT_CONFIRMATION", c.DefaultDisableVisitRedditConfirmation),
			Banner:                             parse("REDLIB_BANNER", c.Banner),
			RobotsDisableIndexing:              parse("REDLIB_ROBOTS_DISABLE_INDEXING", c.RobotsDisableIndexing),
			Pushshift:                          parse("REDLIB_PUSHSHIFT_FRONTEND", c.Pushshift),
			EnableRSS:                          parse("REDLIB_ENABLE_RSS", c.EnableRSS),
			FullURL:                            parse("REDLIB_FULL_URL", c.FullURL),
			DefaultRemoveDefaultFeeds:          parse("REDLIB_DEFAULT_REMOVE_DEFAULT_FEEDS", c.DefaultRemoveDefaultFeeds),
		}
	})
	return globalConfig
}

// GetSetting retrieves a setting by its key.
func GetSetting(name string) *string {
	config := LoadGlobalConfig()
	switch name {
	case "REDLIB_SFW_ONLY":
		return config.SFWOnly
	case "REDLIB_DEFAULT_THEME":
		return config.DefaultTheme
	case "REDLIB_DEFAULT_FRONT_PAGE":
		return config.DefaultFrontPage
	case "REDLIB_DEFAULT_LAYOUT":
		return config.DefaultLayout
	case "REDLIB_DEFAULT_COMMENT_SORT":
		return config.DefaultCommentSort
	case "REDLIB_DEFAULT_POST_SORT":
		return config.DefaultPostSort
	case "REDLIB_DEFAULT_BLUR_SPOILER":
		return config.DefaultBlurSpoiler
	case "REDLIB_DEFAULT_SHOW_NSFW":
		return config.DefaultShowNSFW
	case "REDLIB_DEFAULT_BLUR_NSFW":
		return config.DefaultBlurNSFW
	case "REDLIB_DEFAULT_USE_HLS":
		return config.DefaultUseHLS
	case "REDLIB_DEFAULT_HIDE_HLS_NOTIFICATION":
		return config.DefaultHideHLSNotification
	case "REDLIB_DEFAULT_WIDE":
		return config.DefaultWide
	case "REDLIB_DEFAULT_HIDE_AWARDS":
		return config.DefaultHideAwards
	case "REDLIB_DEFAULT_HIDE_SIDEBAR_AND_SUMMARY":
		return config.DefaultHideSidebarAndSummary
	case "REDLIB_DEFAULT_HIDE_SCORE":
		return config.DefaultHideScore
	case "REDLIB_DEFAULT_SUBSCRIPTIONS":
		return config.DefaultSubscriptions
	case "REDLIB_DEFAULT_FILTERS":
		return config.DefaultFilters
	case "REDLIB_DEFAULT_DISABLE_VISIT_REDDIT_CONFIRMATION":
		return config.DefaultDisableVisitRedditConfirmation
	case "REDLIB_BANNER":
		return config.Banner
	case "REDLIB_ROBOTS_DISABLE_INDEXING":
		return config.RobotsDisableIndexing
	case "REDLIB_PUSHSHIFT_FRONTEND":
		return config.Pushshift
	case "REDLIB_ENABLE_RSS":
		return config.EnableRSS
	case "REDLIB_FULL_URL":
		return config.FullURL
	case "REDLIB_DEFAULT_REMOVE_DEFAULT_FEEDS":
		return config.DefaultRemoveDefaultFeeds
	default:
		return nil
	}
}