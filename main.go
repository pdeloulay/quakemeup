package main

import (
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
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Radius       int     `json:"radius"`
	EnableAlerts bool    `json:"enableAlerts"`
	EnablePush   bool    `json:"enablePush"`
	LastUpdated  string  `json:"lastUpdated"`
}

var (
	userLocations = make(map[string]UserLocation)
	locationMutex = &sync.RWMutex{}
)

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
	TimeAgo   string    `json:"timeAgo"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Depth     float64   `json:"depth"`
	Distance  float64   `json:"distance"`
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
		userID := r.RemoteAddr
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

	// API endpoint for earthquakes
	http.HandleFunc("/api/quakes", quakesHandler)

	// Handle location API
	http.HandleFunc("/api/location", handleLocation)

	// Alert handler
	http.HandleFunc("/alerts", alertHandler)

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

	// Generate a unique ID for the user (in production, use a proper session ID)
	userID := r.RemoteAddr
	loc.LastUpdated = time.Now().Format(time.RFC3339)

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
	φ1 := lat1 * (3.141592653589793 / 180)
	φ2 := lat2 * (3.141592653589793 / 180)
	Δφ := (lat2 - lat1) * (3.141592653589793 / 180)
	Δλ := (lon2 - lon1) * (3.141592653589793 / 180)

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

	// Get earthquakes from the last 24 hours
	startTime := time.Now().Add(-24 * time.Hour)
	earthquakes, err := fetchEarthquakes(startTime, 0, 0)
	if err != nil {
		log.Printf("Error in quakesHandler: %v", err)
		http.Error(w, "Failed to fetch earthquake data", http.StatusInternalServerError)
		return
	}

	// Filter by user location if provided
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")
	if lat != "" && lon != "" {
		userLat, _ := strconv.ParseFloat(lat, 64)
		userLon, _ := strconv.ParseFloat(lon, 64)

		var filteredQuakes []Earthquake
		for _, quake := range earthquakes {
			distance := calculateDistance(userLat, userLon, quake.Latitude, quake.Longitude)
			if distance <= 500 { // 500km radius
				filteredQuakes = append(filteredQuakes, quake)
			}
		}
		earthquakes = filteredQuakes
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(earthquakes)
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
	tmpl := template.Must(template.ParseFiles("templates/alert.html"))
	err = tmpl.Execute(w, latestQuake)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
