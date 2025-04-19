package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
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

// Earthquake represents a single earthquake event
type Earthquake struct {
	Time      int64   `json:"time"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Magnitude float64 `json:"mag"`
	Place     string  `json:"place"`
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

// Test data for development
var testEarthquakes = []Earthquake{
	{
		Time:      time.Now().Unix() * 1000,
		Latitude:  37.7749,
		Longitude: -122.4194,
		Magnitude: 4.5,
		Place:     "San Francisco, California",
	},
	{
		Time:      time.Now().Unix() * 1000,
		Latitude:  34.0522,
		Longitude: -118.2437,
		Magnitude: 3.2,
		Place:     "Los Angeles, California",
	},
	{
		Time:      time.Now().Unix() * 1000,
		Latitude:  40.7128,
		Longitude: -74.0060,
		Magnitude: 2.8,
		Place:     "New York, New York",
	},
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
	http.HandleFunc("/api/quakes", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request for earthquake data from %s", r.RemoteAddr)

		// Get user's location from stored data
		userID := r.RemoteAddr
		locationMutex.RLock()
		userLoc, hasLocation := userLocations[userID]
		locationMutex.RUnlock()

		var earthquakes []Earthquake

		// In development, use test data
		if os.Getenv("ENV") == "development" {
			log.Printf("Development mode: Using test earthquake data")
			earthquakes = testEarthquakes
		} else {
			// Fetch earthquakes from USGS API
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			req, err := http.NewRequest("GET", "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson", nil)
			if err != nil {
				log.Printf("Error creating request: %v", err)
				http.Error(w, "Error creating request", http.StatusInternalServerError)
				return
			}

			req.Header.Set("User-Agent", "QuakeMeUp/1.0")

			log.Printf("Fetching earthquake data from USGS API")
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error fetching earthquakes: %v", err)
				http.Error(w, "Error fetching earthquakes", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("USGS API returned non-200 status: %d", resp.StatusCode)
				http.Error(w, "Error fetching earthquakes", http.StatusInternalServerError)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response: %v", err)
				http.Error(w, "Error reading response", http.StatusInternalServerError)
				return
			}

			var usgsResponse USGSResponse
			if err := json.Unmarshal(body, &usgsResponse); err != nil {
				log.Printf("Error parsing response: %v", err)
				http.Error(w, "Error parsing response", http.StatusInternalServerError)
				return
			}

			// Convert USGS response to our Earthquake format
			earthquakes = make([]Earthquake, len(usgsResponse.Features))
			for i, feature := range usgsResponse.Features {
				earthquakes[i] = Earthquake{
					Time:      feature.Properties.Time,
					Latitude:  feature.Geometry.Coordinates[1],
					Longitude: feature.Geometry.Coordinates[0],
					Magnitude: feature.Properties.Magnitude,
					Place:     feature.Properties.Place,
				}
			}
		}

		// Filter earthquakes based on user's location if available
		if hasLocation {
			log.Printf("Filtering earthquakes for user %s within %d km radius", userID, userLoc.Radius)
			filteredEarthquakes := make([]Earthquake, 0)
			for _, quake := range earthquakes {
				distance := calculateDistance(userLoc.Latitude, userLoc.Longitude, quake.Latitude, quake.Longitude)
				if distance <= float64(userLoc.Radius) {
					filteredEarthquakes = append(filteredEarthquakes, quake)
				}
			}
			earthquakes = filteredEarthquakes
			log.Printf("Found %d earthquakes within user's radius", len(earthquakes))
		} else {
			log.Printf("No location data available, returning all earthquakes")
		}

		log.Printf("Returning %d earthquakes", len(earthquakes))

		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Send response
		if err := json.NewEncoder(w).Encode(earthquakes); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	})

	// Handle location API
	http.HandleFunc("/api/location", handleLocation)

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
