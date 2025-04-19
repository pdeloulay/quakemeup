# QuakeMeUp

Real-time earthquake monitoring and alert system that keeps you informed about seismic activities worldwide.

## CSS Palette

#3D0814
#E7F9A9
#C6B38E
#9A9B73
#442F38

## Features

- Real-time earthquake monitoring
- Location-based alerts
- Customizable alert radius (10-500 km)
- Push notifications
- Mobile-friendly interface
- Interactive world map visualization
- Geolocation support
- Interactive world map showing recent earthquakes
- List of recent earthquakes with magnitude and location
- Automatic data refresh every 5 minutes

## Changelog

### [0.4.8] - 2024-03-19
- Enhanced API data validation and error handling
- Added detailed debug logging for earthquake data
- Improved magnitude and location data processing
- Added strict type checking for earthquake properties
- Enhanced error messages for invalid data
- Added proper HTTP status code validation

### [0.4.7] - 2024-03-19
- Added robust error handling for earthquake data
- Improved handling of undefined or invalid magnitude values
- Added user-friendly error messages for API failures
- Enhanced data validation for earthquake objects
- Added fallback values for missing earthquake properties
- Improved error logging and debugging information

### [0.4.6] - 2024-03-19
- Added detailed tooltip overlay for earthquake markers
- Enhanced marker hover interactions with smooth animations
- Added comprehensive earthquake information display
- Improved map interactivity with non-intrusive tooltips
- Updated styling to match color palette
- Added depth information to earthquake details

### [0.4.5] - 2024-03-19
- Updated radius slider to use half Earth's circumference as maximum (20,037 km)
- Added radius persistence in localStorage
- Enhanced radius display with thousands separators
- Added explanatory text for maximum radius
- Improved radius handling in location API integration
- Added radius restoration from saved location data

### [0.4.4] - 2024-03-19
- Updated earthquake proximity detection to use user-specified radius
- Enhanced logging for radius-based user filtering
- Added detailed logging for alert status and distance calculations
- Improved user proximity detection accuracy
- Added logging for disabled alerts

### [0.4.3] - 2024-03-19
- Added new `/api/quakes` endpoint for getting users near earthquakes
- Enhanced earthquake proximity detection with detailed logging
- Added response structure for nearby users
- Improved error handling for API requests
- Added distance calculation logging

### [0.4.2] - 2024-03-19
- Added user location filtering for earthquakes
- Enhanced map page with user location marker
- Added user location pulse animation
- Updated earthquake list title to show radius when location is available
- Improved logging for map page and API requests
- Added user location data to map template

### [0.4.1] - 2024-03-19
- Enhanced USGS API integration with better error handling
- Added development mode with test earthquake data
- Improved logging for API requests and responses
- Added timeout and user-agent headers for API requests
- Added status code validation for API responses

### [0.4.0] - 2024-03-19
- Added interactive map page showing recent earthquakes
- Integrated USGS API for real-time earthquake data
- Added earthquake list with magnitude and location details
- Implemented automatic data refresh every 5 minutes
- Enhanced world map visualization with preferred color palette
- Added earthquake markers with magnitude-based styling
- Implemented responsive earthquake list with scrollable view
- Added hover effects and animations for better user experience

### [0.3.0] - 2024-03-19
- Added geolocation support
- Implemented location-based alert system
- Added user preferences for alert radius
- Added push notification support
- Added location settings panel
- Implemented distance calculation for earthquake alerts
- Added user location storage and management
- Enhanced security with mutex for concurrent access

### [0.2.0] - 2024-03-19
- Enhanced mobile responsiveness
- Added mobile-first design approach
- Implemented hamburger menu for mobile navigation
- Improved touch interactions and accessibility
- Optimized layouts for different screen sizes
- Added SVG-based world map background
- Enhanced earthquake visualization animations

### [0.1.0] - 2024-03-19
- Initial project setup
- Added landing page with responsive design
- Implemented basic Go server
- Added world map background with earthquake pulse animation
- Integrated Tailwind CSS for styling

## Development

### Prerequisites
- Go 1.16 or higher
- Modern web browser with geolocation support
- USGS Earthquake API

### Running the Application
1. Clone the repository
2. Run `go run main.go`
3. Visit `http://localhost:8080` in your browser

### API Endpoints
- `GET /` - Main landing page
- `POST /api/location` - Store user location and preferences
  - Request body:
    ```json
    {
      "latitude": float64,
      "longitude": float64,
      "radius": int,
      "enableAlerts": bool,
      "enablePush": bool
    }
    ```

## Security Considerations
- User locations are stored in memory (in production, use a proper database)
- Geolocation data is only used for alert purposes
- Push notifications require explicit user permission
- All sensitive operations are protected with mutex locks

## Mobile Support
The application is fully responsive and optimized for mobile devices. Key features include:
- Touch-friendly navigation
- Responsive layouts
- Optimized performance
- Mobile-first design approach
- Geolocation support
- Push notifications

## Contributing
1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License
This project is licensed under the MIT License - see the LICENSE file for details.
