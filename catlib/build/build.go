package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Get the git hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		// If git fails, use a default value
		out = []byte("dev")
	}

	// Create the version file
	f, err := os.Create("version.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Write the git hash to the version file
	fmt.Fprintf(f, "package main\n\nconst GitHash = \"%s\"\n", strings.TrimSpace(string(out)))
}