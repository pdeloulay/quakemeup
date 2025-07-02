package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type PageData struct {
	Title string
}

type UserLocation struct {
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Radius       int       `json:"radius"`
	EnableAlerts bool      `json:"enableAlerts"`
	EnablePush   bool      `json:"enablePush"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

var (
	userLocations = make(map[string]UserLocation)
	locationMutex = &sync.RWMutex{}
	sessions      = make(map[string]string) // sessionID -> userID mapping
	sessionMutex  = &sync.RWMutex{}
)

// generateSessionID creates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getUserID extracts or creates a user session
func getUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// No session cookie, create a new one
		sessionID, err := generateSessionID()
		if err != nil {
			return "", err
		}

		sessionMutex.Lock()
		sessions[sessionID] = sessionID // Use sessionID as userID for simplicity
		sessionMutex.Unlock()

		return sessionID, nil
	}

	sessionMutex.RLock()
	userID, exists := sessions[cookie.Value]
	sessionMutex.RUnlock()

	if !exists {
		// Invalid session, create new one
		sessionID, err := generateSessionID()
		if err != nil {
			return "", err
		}

		sessionMutex.Lock()
		sessions[sessionID] = sessionID
		sessionMutex.Unlock()

		return sessionID, nil
	}

	return userID, nil
}

// setSessionCookie sets a secure session cookie
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// cleanupOldData removes stale sessions and location data older than 7 days
func cleanupOldData() {
	ticker := time.NewTicker(24 * time.Hour) // Run cleanup daily
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-7 * 24 * time.Hour) // 7 days ago

		locationMutex.Lock()
		for userID, location := range userLocations {
			if location.LastUpdated.Before(cutoff) {
				delete(userLocations, userID)
				log.Printf("Cleaned up old location data for user %s", userID)
			}
		}
		locationMutex.Unlock()

		// Note: Sessions cleanup would need more sophisticated tracking
		// of last access time. For now, we rely on cookie expiration.
		log.Printf("Completed cleanup cycle, removed stale data older than %v", cutoff)
	}
}

// FeatureCollection represents the USGS GeoJSON response
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
}

type Properties struct {
	Magnitude float64 `json:"mag"`
	Place     string  `json:"place"`
	Time      int64   `json:"time"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Earthquake struct {
	Magnitude float64   `json:"magnitude"`
	Place     string    `json:"place"`
	Time      time.Time `json:"time"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Depth     float64   `json:"depth"`
	TimeAgo   string    `json:"timeAgo"`
	Distance  float64   `json:"distance,omitempty"` // Distance from user in km, if available
}

// USGSResponse represents the response from USGS API
type USGSResponse struct {
	Features []struct {
		Properties struct {
			Time      int64   `json:"time"`
			Magnitude float64 `json:"mag"`
			Place     string  `json:"place"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

// AlertPageData represents the data for the alert page template
type AlertPageData struct {
	Title       string
	LatestQuake *Earthquake
}

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/home.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Map page
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Map page requested from %s", r.RemoteAddr)

		// Get user's location from stored data
		userID, err := getUserID(r)
		if err != nil {
			log.Printf("Error getting user ID: %v", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}
		setSessionCookie(w, userID)

		locationMutex.RLock()
		userLoc, hasLocation := userLocations[userID]
		locationMutex.RUnlock()

		// Prepare template data
		data := struct {
			HasLocation bool
			UserLat     float64
			UserLng     float64
			Radius      int
		}{
			HasLocation: hasLocation,
		}

		if hasLocation {
			log.Printf("User %s has location data: %+v", userID, userLoc)
			data.UserLat = userLoc.Latitude
			data.UserLng = userLoc.Longitude
			data.Radius = userLoc.Radius
		} else {
			log.Printf("No location data found for user %s", userID)
		}

		tmpl, err := template.ParseFiles("templates/map.html")
		if err != nil {
			log.Printf("Error parsing map template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error executing map template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Privacy Policy page
	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/privacy.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// Terms of Service page
	http.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/terms.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// About page
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/about.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	// API endpoint for earthquakes
	http.HandleFunc("/api/quakes", quakesHandler)

	// Handle location API
	http.HandleFunc("/api/location", handleLocation)

	// Alert handler
	http.HandleFunc("/latest", alertHandler)

	// Start cleanup goroutine
	go cleanupOldData()

	// Start server
	port := 8080
	log.Printf("Server starting on port %d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handleLocation(w http.ResponseWriter, r *http.Request) {
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

	// Get secure user ID from session
	userID, err := getUserID(r)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	setSessionCookie(w, userID)

	loc.LastUpdated = time.Now()

	// Store the location
	locationMutex.Lock()
	userLocations[userID] = loc
	locationMutex.Unlock()

	log.Printf("Stored location for user %s: %+v", userID, loc)

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Location stored successfully",
	})
}

// Simplified distance calculation (Haversine formula)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	// Convert degrees to radians using math.Pi for better precision
	φ1 := lat1 * (math.Pi / 180)
	φ2 := lat2 * (math.Pi / 180)
	Δφ := (lat2 - lat1) * (math.Pi / 180)
	Δλ := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// timeAgo returns a human-readable string representing how long ago a time was
func timeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// fetchEarthquakes retrieves earthquake data from USGS API with optional filtering
func fetchEarthquakes(startTime time.Time, minMagnitude float64, limit int) ([]Earthquake, error) {
	log.Printf("Fetching earthquakes since %v with minimum magnitude %v and limit %v", startTime, minMagnitude, limit)

	// Construct USGS API URL with parameters
	baseURL := "https://earthquake.usgs.gov/fdsnws/event/1/query"
	params := url.Values{}
	params.Add("format", "geojson")
	params.Add("starttime", startTime.Format(time.RFC3339))
	params.Add("minmagnitude", fmt.Sprintf("%.1f", minMagnitude))
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	apiURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	log.Printf("USGS API URL: %s", apiURL)

	// Make HTTP request
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Error fetching earthquake data: %v", err)
		return nil, fmt.Errorf("failed to fetch earthquake data: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result FeatureCollection
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Convert features to earthquakes
	earthquakes := make([]Earthquake, 0, len(result.Features))
	for _, feature := range result.Features {
		quakeTime := time.Unix(feature.Properties.Time/1000, 0)
		earthquake := Earthquake{
			Magnitude: feature.Properties.Magnitude,
			Place:     feature.Properties.Place,
			Time:      quakeTime,
			TimeAgo:   timeAgo(quakeTime),
			Latitude:  feature.Geometry.Coordinates[1],
			Longitude: feature.Geometry.Coordinates[0],
			Depth:     feature.Geometry.Coordinates[2],
			Distance:  0,
		}
		earthquakes = append(earthquakes, earthquake)
	}

	log.Printf("Successfully fetched %d earthquakes", len(earthquakes))
	return earthquakes, nil
}

func quakesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /api/quakes request")

	// Get user's location from stored data
	userID, err := getUserID(r)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	locationMutex.RLock()
	userLoc, hasLocation := userLocations[userID]
	locationMutex.RUnlock()

	// Parse query parameters
	query := r.URL.Query()
	duration := query.Get("duration")
	minMagStr := query.Get("minMagnitude")
	maxMagStr := query.Get("maxMagnitude")
	radiusStr := query.Get("radius")

	// Set default values
	var startTime time.Time
	switch duration {
	case "hour":
		startTime = time.Now().Add(-time.Hour)
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default: // default to 24 hours
		startTime = time.Now().AddDate(0, 0, -1)
	}

	minMag := 0.0
	if minMagStr != "" {
		minMag, _ = strconv.ParseFloat(minMagStr, 64)
	}

	maxMag := 10.0
	if maxMagStr != "" {
		maxMag, _ = strconv.ParseFloat(maxMagStr, 64)
	}

	radius := 100 // default radius in km
	if radiusStr != "" {
		radius, _ = strconv.Atoi(radiusStr)
	}

	// Fetch earthquakes from USGS
	earthquakes, err := fetchEarthquakes(startTime, minMag, 500) // increased limit to account for filtering
	if err != nil {
		log.Printf("Error fetching earthquakes: %v", err)
		http.Error(w, "Failed to fetch earthquake data", http.StatusInternalServerError)
		return
	}

	// Filter and process earthquakes
	filteredQuakes := make([]Earthquake, 0)
	for _, quake := range earthquakes {
		// Apply magnitude filter
		if quake.Magnitude < minMag || quake.Magnitude > maxMag {
			continue
		}

		// Calculate distance if user location is available
		if hasLocation {
			distance := calculateDistance(userLoc.Latitude, userLoc.Longitude, quake.Latitude, quake.Longitude)
			quake.Distance = distance // Add distance to earthquake data

			// Apply radius filter
			if distance > float64(radius) {
				continue
			}
		}

		filteredQuakes = append(filteredQuakes, quake)
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredQuakes)
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /api/alerts request")

	// Get the latest earthquake
	startTime := time.Now().Add(-24 * time.Hour)
	earthquakes, err := fetchEarthquakes(startTime, 0, 1)
	if err != nil {
		log.Printf("Error in alertHandler: %v", err)
		http.Error(w, "Failed to fetch earthquake data", http.StatusInternalServerError)
		return
	}

	if len(earthquakes) == 0 {
		log.Println("No earthquakes found in the last 24 hours")
		http.Error(w, "No earthquakes found", http.StatusNotFound)
		return
	}

	latestQuake := earthquakes[0]

	// Render alert template
	tmpl := template.Must(template.ParseFiles("templates/latest.html"))
	err = tmpl.Execute(w, latestQuake)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
