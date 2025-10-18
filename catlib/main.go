package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	// Command-line flags
	addr := flag.String("address", ":8080", "Address and port to listen on")
	flag.Parse()

	// Initialize modules
	initializeOAuth()
	LoadGlobalConfig()
	GetInstanceInfo()

	// --- ROUTING ---
	// Using a simple ServeMux for now. A more robust router would be better for production.
	mux := http.NewServeMux()

	// Static files
	mux.HandleFunc("/static/", staticFileHandler)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	// API/proxy routes (simplified)
	// In a real app, these would be more robustly handled.
	mux.HandleFunc("/vid/", proxyHandler("https://v.redd.it"))
	mux.HandleFunc("/img/", proxyHandler("https://i.redd.it"))

	// Page handlers
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/r/", subredditRouteHandler)
	mux.HandleFunc("/user/", userRouteHandler)
	mux.HandleFunc("/settings", settingsGetHandler)
	mux.HandleFunc("/settings/set", settingsSetHandler)
	mux.HandleFunc("/settings/restore", settingsRestoreHandler)
	mux.HandleFunc("/search", findHandler)
	mux.HandleFunc("/info", instanceInfoHandler)
	mux.HandleFunc("/info/", instanceInfoHandler) // Handle .ext

	// Start server
	log.Printf("Starting Catlib server on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// If path is not just "/", it's a 404 or a special root path like /hot, /new etc.
	if r.URL.Path != "/" {
		// Handle /hot, /new, etc. - pass to community handler
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) == 1 && isSortType(parts[0]) {
			// This is a sort type on the front page
			communityHandler(w, r)
			return
		}
		http.NotFound(w, r)
		return
	}
	communityHandler(w, r)
}

func subredditRouteHandler(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/r/"), "/")
    sub := parts[0]
    r = r.WithContext(context.WithValue(r.Context(), "sub", sub))

    if len(parts) == 1 {
        communityHandler(w, r)
        return
    }

    // Handle sub-routes like /r/{sub}/comments/...
    switch parts[1] {
    case "comments":
        // This is a simplified way to extract further params
        if len(parts) > 2 {
            r = r.WithContext(context.WithValue(r.Context(), "id", parts[2]))
        }
        postItem(w, r)
    case "duplicates":
        if len(parts) > 2 {
            r = r.WithContext(context.WithValue(r.Context(), "id", parts[2]))
        }
        duplicatesItem(w, r)
    case "wiki", "about":
        // Placeholder for wiki/sidebar handlers
        http.NotFound(w, r)
    default:
        // Assume it's a sort type, e.g., /r/popular/hot
        r = r.WithContext(context.WithValue(r.Context(), "sort", parts[1]))
        communityHandler(w, r)
    }
}

func userRouteHandler(w http.ResponseWriter, r *http.Request) {
	// /user/{name}/{listing}
	profileHandler(w, r)
}

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	// Be careful with this in production!
	http.ServeFile(w, r, r.URL.Path[1:])
}

func proxyHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simplified proxy logic
		// A real implementation would handle headers, etc.
		url := fmt.Sprintf("%s%s", target, r.URL.Path)
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func isSortType(s string) bool {
	switch s {
	case "best", "hot", "new", "top", "rising", "controversial":
		return true
	default:
		return false
	}
}