package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// Placeholders for build-time variables
var (
	PackageName  = "catlib"
	CrateVersion = "0.1.0"
	GitCommit    = "unknown"
	CompileMode  = "Debug"
)

// InstanceInfo holds information about the running instance.
type InstanceInfo struct {
	PackageName   string  `json:"package_name" yaml:"package_name"`
	CrateVersion  string  `json:"crate_version" yaml:"crate_version"`
	GitCommit     string  `json:"git_commit" yaml:"git_commit"`
	DeployDate    string  `json:"deploy_date" yaml:"deploy_date"`
	CompileMode   string  `json:"compile_mode" yaml:"compile_mode"`
	DeployUnixTS  int64   `json:"deploy_unix_ts" yaml:"deploy_unix_ts"`
	Config        *Config `json:"config" yaml:"config"`
}

var (
	instanceInfo     *InstanceInfo
	instanceInfoOnce sync.Once
)

// GetInstanceInfo creates and returns a singleton InstanceInfo object.
func GetInstanceInfo() *InstanceInfo {
	instanceInfoOnce.Do(func() {
		instanceInfo = &InstanceInfo{
			PackageName:  PackageName,
			CrateVersion: CrateVersion,
			GitCommit:    GitCommit,
			DeployDate:   time.Now().String(),
			CompileMode:  CompileMode,
			DeployUnixTS: time.Now().Unix(),
			Config:       LoadGlobalConfig(),
		}
	})
	return instanceInfo
}

// instanceInfoHandler handles requests to the instance info endpoint.
func instanceInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Assumes a router that extracts the extension.
	parts := strings.Split(r.URL.Path, ".")
	extension := ""
	if len(parts) > 1 {
		extension = parts[len(parts)-1]
	}

	info := GetInstanceInfo()

	switch extension {
	case "yaml", "yml":
		infoYAML(w, r, info)
	case "txt":
		infoTxt(w, r, info)
	case "json":
		infoJSON(w, r, info)
	case "html", "":
		infoHTML(w, r, info)
	default:
		http.Error(w, "Error: Invalid info extension", http.StatusNotFound)
	}
}

func infoJSON(w http.ResponseWriter, r *http.Request, info *InstanceInfo) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Error serializing JSON", http.StatusInternalServerError)
	}
}

func infoYAML(w http.ResponseWriter, r *http.Request, info *InstanceInfo) {
	w.Header().Set("Content-Type", "application/yaml")
	if err := yaml.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Error serializing YAML", http.StatusInternalServerError)
	}
}

func infoTxt(w http.ResponseWriter, r *http.Request, info *InstanceInfo) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, info.ToString(stringTypeRaw))
}

func infoHTML(w http.ResponseWriter, r *http.Request, info *InstanceInfo) {
	// This is a placeholder for a real template.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Using a simple message template placeholder
	body := info.ToString(stringTypeHTML)
	// Assuming a MessageTemplate struct and render method exist for simplicity
	fmt.Fprintf(w, "<h1>Instance Information</h1>%s", body)
}

type stringType int

const (
	stringTypeRaw stringType = iota
	stringTypeHTML
)

func (info *InstanceInfo) ToString(stype stringType) string {
	if stype == stringTypeHTML {
		return info.toTable()
	}

	// Raw string format
	var b strings.Builder
	fmt.Fprintf(&b, "Package name: %s\n", info.PackageName)
	fmt.Fprintf(&b, "Crate version: %s\n", info.CrateVersion)
	fmt.Fprintf(&b, "Git commit: %s\n", info.GitCommit)
	fmt.Fprintf(&b, "Deploy date: %s\n", info.DeployDate)
	fmt.Fprintf(&b, "Deploy timestamp: %d\n", info.DeployUnixTS)
	fmt.Fprintf(&b, "Compile mode: %s\n", info.CompileMode)
	// Omitting config details for brevity in raw text, similar to rust impl
	return b.String()
}

func (info *InstanceInfo) toTable() string {
	// Manual HTML table generation
	var b strings.Builder
	convert := func(s *string) string {
		if s == nil {
			return `<span class="unset"><i>Unset</i></span>`
		}
		return *s
	}

	if info.Config.Banner != nil {
		fmt.Fprintf(&b, "<h3>Instance banner</h3><br /><p>%s</p><br />", *info.Config.Banner)
	}

	b.WriteString(`<table border="1"><tr><th colspan="2">Settings</th></tr>`)
	fmt.Fprintf(&b, "<tr><td>Package name</td><td>%s</td></tr>", info.PackageName)
	fmt.Fprintf(&b, "<tr><td>Crate version</td><td>%s</td></tr>", info.CrateVersion)
	fmt.Fprintf(&b, "<tr><td>Git commit</td><td>%s</td></tr>", info.GitCommit)
	fmt.Fprintf(&b, "<tr><td>Deploy date</td><td>%s</td></tr>", info.DeployDate)
	fmt.Fprintf(&b, "<tr><td>Deploy timestamp</td><td>%d</td></tr>", info.DeployUnixTS)
	fmt.Fprintf(&b, "<tr><td>Compile mode</td><td>%s</td></tr>", info.CompileMode)
	fmt.Fprintf(&b, "<tr><td>SFW only</td><td>%s</td></tr>", convert(info.Config.SFWOnly))
	fmt.Fprintf(&b, "<tr><td>Pushshift frontend</td><td>%s</td></tr>", convert(info.Config.Pushshift))
	fmt.Fprintf(&b, "<tr><td>RSS enabled</td><td>%s</td></tr>", convert(info.Config.EnableRSS))
	fmt.Fprintf(&b, "<tr><td>Full URL</td><td>%s</td></tr>", convert(info.Config.FullURL))
	fmt.Fprintf(&b, "<tr><td>Remove default feeds</td><td>%s</td></tr>", convert(info.Config.DefaultRemoveDefaultFeeds))
	b.WriteString("</table><br />")

	b.WriteString(`<table border="1"><tr><th colspan="2">Default preferences</th></tr>`)
	fmt.Fprintf(&b, "<tr><td>Hide awards</td><td>%s</td></tr>", convert(info.Config.DefaultHideAwards))
	fmt.Fprintf(&b, "<tr><td>Hide score</td><td>%s</td></tr>", convert(info.Config.DefaultHideScore))
	// ... add all other preferences ...
	b.WriteString("</table>")

	return b.String()
}