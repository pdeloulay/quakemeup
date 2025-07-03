package api

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// App represents the application structure
type App struct {
	templates     *template.Template
	userLocations map[string]UserLocation
	locationMutex *sync.RWMutex
	sessions      map[string]string
	sessionMutex  *sync.RWMutex
	mux           *http.ServeMux
}

// NewApp creates and initializes a new App instance
func NewApp() *App {
	app := &App{
		userLocations: make(map[string]UserLocation),
		locationMutex: &sync.RWMutex{},
		sessions:      make(map[string]string),
		sessionMutex:  &sync.RWMutex{},
		mux:           http.NewServeMux(),
	}

	// Load environment variables
	if err := loadEnv(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Read version
	version := readVersion()
	log.Printf("Starting QuakeMeUp version %s", version)

	// Parse templates
	templates, err := template.ParseFiles(
		"templates/home.html",
		"templates/about.html",
		"templates/latest.html",
		"templates/mapgl.html",
		"templates/mission.html",
		"templates/privacy.html",
		"templates/terms.html",
	)
	if err != nil {
		log.Printf("Warning: Error parsing templates: %v", err)
	}
	app.templates = templates

	// Setup routes
	app.setupRoutes()

	return app
}

// ServeHTTP implements the http.Handler interface
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.mux.ServeHTTP(w, r)
}

// setupRoutes configures all the application routes
func (app *App) setupRoutes() {
	// Serve static files
	app.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Setup all routes
	app.mux.HandleFunc("/", app.homeHandler)
	app.mux.HandleFunc("/about", app.aboutHandler)
	app.mux.HandleFunc("/mapgl", app.mapHandler)
	app.mux.HandleFunc("/latest", app.latestHandler)
	app.mux.HandleFunc("/privacy", app.privacyHandler)
	app.mux.HandleFunc("/terms", app.termsHandler)
	app.mux.HandleFunc("/api/quakes", app.quakesHandler)
	app.mux.HandleFunc("/api/location", app.locationHandler)
}

// loadEnv loads environment variables from .env file
func loadEnv() error {
	// Read .env file
	data, err := os.ReadFile(".env")
	if err != nil {
		return err
	}

	// Parse and set environment variables
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`) // Remove quotes if present

		os.Setenv(key, value)
	}

	return nil
}

// readVersion reads the version from .version file
func readVersion() string {
	version, err := os.ReadFile(".version")
	if err != nil {
		log.Printf("Warning: Could not read version file: %v", err)
		return "dev"
	}
	return strings.TrimSpace(string(version))
}
