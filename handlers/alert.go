package handlers

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/pdeloulay/quakemeup/models"
)

// haversineDistance calculates the distance between two points on Earth using the Haversine formula
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers
	lat1Rad := lat1 * (3.141592653589793 / 180)
	lon1Rad := lon1 * (3.141592653589793 / 180)
	lat2Rad := lat2 * (3.141592653589793 / 180)
	lon2Rad := lon2 * (3.141592653589793 / 180)

	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := (dlat/2)*(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*(dlon/2)*(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}

// AlertHandler handles the alert page
func AlertHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AlertHandler: Processing request")

	// Get user's location from query parameters
	userLat := r.URL.Query().Get("lat")
	userLon := r.URL.Query().Get("lon")

	// Get the latest earthquake
	earthquake, err := models.GetLatestEarthquake()
	if err != nil {
		log.Printf("AlertHandler: Error getting latest earthquake: %v", err)
		http.Error(w, "Error getting earthquake data", http.StatusInternalServerError)
		return
	}

	// Calculate distance if user location is available
	var distance float64
	if userLat != "" && userLon != "" {
		userLatFloat, err := strconv.ParseFloat(userLat, 64)
		if err != nil {
			log.Printf("AlertHandler: Error parsing user latitude: %v", err)
		}
		userLonFloat, err := strconv.ParseFloat(userLon, 64)
		if err != nil {
			log.Printf("AlertHandler: Error parsing user longitude: %v", err)
		}
		if err == nil {
			distance = haversineDistance(userLatFloat, userLonFloat, earthquake.Latitude, earthquake.Longitude)
		}
	}

	// Prepare the template data
	data := struct {
		Magnitude float64
		Place     string
		Latitude  float64
		Longitude float64
		TimeAgo   string
		Depth     float64
		Distance  float64
	}{
		Magnitude: earthquake.Magnitude,
		Place:     earthquake.Place,
		Latitude:  earthquake.Latitude,
		Longitude: earthquake.Longitude,
		TimeAgo:   time.Since(earthquake.Time).Round(time.Second).String(),
		Depth:     earthquake.Depth,
		Distance:  distance,
	}

	// Render the template
	err = templates.RenderTemplate(w, "alert.html", data)
	if err != nil {
		log.Printf("AlertHandler: Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

	log.Println("AlertHandler: Request processed successfully")
}
