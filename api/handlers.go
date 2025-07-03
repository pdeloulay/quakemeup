package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// homeHandler serves the home page
func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:   "QuakeMeUp - Real-time Earthquake Monitoring",
		Version: readVersion(),
	}
	if err := app.templates.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// mapHandler serves the map page
func (app *App) mapHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserID(r)
	app.sessionMutex.RLock()
	_, hasSession := app.sessions[userID]
	app.sessionMutex.RUnlock()

	if hasSession {
		setSessionCookie(w, userID)
	}

	app.locationMutex.RLock()
	userLoc, hasLocation := app.userLocations[userID]
	app.locationMutex.RUnlock()

	data := MapPageData{
		HasLocation: hasLocation,
		MapboxToken: os.Getenv("MAPBOX_TOKEN"),
		Version:     readVersion(),
	}

	if hasLocation {
		data.UserLat = userLoc.Latitude
		data.UserLng = userLoc.Longitude
		data.Radius = userLoc.Radius
	}

	if err := app.templates.ExecuteTemplate(w, "mapgl.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// aboutHandler serves the about page
func (app *App) aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "About - QuakeMeUp",
		Version: readVersion(),
	}
	if err := app.templates.ExecuteTemplate(w, "about.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// latestHandler serves the latest earthquakes page
func (app *App) latestHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Latest Earthquakes - QuakeMeUp",
		Version: readVersion(),
	}
	if err := app.templates.ExecuteTemplate(w, "latest.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// privacyHandler serves the privacy policy page
func (app *App) privacyHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Privacy Policy - QuakeMeUp",
		Version: readVersion(),
	}
	if err := app.templates.ExecuteTemplate(w, "privacy.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// termsHandler serves the terms of service page
func (app *App) termsHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Terms of Service - QuakeMeUp",
		Version: readVersion(),
	}
	if err := app.templates.ExecuteTemplate(w, "terms.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// locationHandler handles user location updates
func (app *App) locationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loc UserLocation
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		log.Printf("Error decoding location: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	setSessionCookie(w, userID)

	loc.LastUpdated = time.Now()

	app.locationMutex.Lock()
	app.userLocations[userID] = loc
	app.locationMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Location stored successfully",
	})
}

// quakesHandler serves earthquake data
func (app *App) quakesHandler(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder - implement actual USGS API integration
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]Earthquake{})
}
