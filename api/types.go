package api

import "time"

// UserLocation represents a user's location data
type UserLocation struct {
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Radius       int       `json:"radius"`
	EnableAlerts bool      `json:"enableAlerts"`
	EnablePush   bool      `json:"enablePush"`
	LastUpdated  time.Time `json:"lastUpdated"`
}

// PageData represents the basic page template data
type PageData struct {
	Title   string
	Version string
}

// MapPageData represents the data for the map page template
type MapPageData struct {
	HasLocation bool
	UserLat     float64
	UserLng     float64
	Radius      int
	MapboxToken string
	Version     string
}

// Earthquake represents earthquake data
type Earthquake struct {
	Magnitude float64   `json:"magnitude"`
	Place     string    `json:"place"`
	Time      time.Time `json:"time"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Depth     float64   `json:"depth"`
	TimeAgo   string    `json:"timeAgo"`
	Distance  float64   `json:"distance,omitempty"`
}
