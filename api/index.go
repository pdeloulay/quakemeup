package api

import (
	"net/http"
	"os"
	"path/filepath"
)

var (
	// Ensure templates and static files are found relative to the api directory
	projectRoot = filepath.Join(filepath.Dir(filepath.Dir(os.Args[0])))
)

// Handler - Vercel serverless function handler
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set working directory to project root
	os.Chdir(projectRoot)

	// Initialize the application
	app := NewApp()

	// Handle the request
	app.ServeHTTP(w, r)
} 