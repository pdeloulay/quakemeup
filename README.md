# QuakeMeUp

Real-time earthquake monitoring and alert system that keeps you informed about seismic activities worldwide.
Hosted at https://quakemeup.leapcell.app/ 

## Table of Contents
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Environment Setup](#environment-setup)
  - [Installation](#installation)
  - [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [Mobile Support](#mobile-support)
- [Development](#development)
  - [Color Palette](#color-palette)
  - [Security Considerations](#security-considerations)
  - [Contributing](#contributing)
- [Deployment](#deployment)
  - [Vercel Deployment](#vercel-deployment)
- [Changelog](#changelog)
- [License](#license)

## Features

- Real-time earthquake monitoring
- Location-based alerts
- Mobile-friendly interface
- Interactive world map visualization
- Geolocation support
- Interactive world map showing recent earthquakes

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Modern web browser with geolocation support
- USGS Earthquake API access
- Git

### Environment Setup

1. Create a `.env` file in the root directory:
```env
# Mapbox Configuration (required for map visualization)
MAPBOX_TOKEN=your_mapbox_token_here
```

### Installation (local)

1. Clone the repository:
```bash
git clone https://github.com/yourusername/quakemeup.git
cd quakemeup
```

2. Install dependencies:
```bash
go mod download
```

### Running the Application

1. Start the server:
```bash
go run main.go
```

2. Visit `http://localhost:8080` in your browser

## API Documentation

### Endpoints

#### Page Routes
- `GET /` - Main landing page
- `GET /about` - About page with mission statement and partner organizations
- `GET /mapgl` - Interactive map page with real-time earthquake visualization
- `GET /latest` - Latest earthquake alerts page
- `GET /privacy` - Privacy policy page
- `GET /terms` - Terms of service page

#### API Routes

##### Location API
- `POST /api/location` - Store user location and preferences
  - Request body:
    ```json
    {
      "latitude": float64,
      "longitude": float64,
      "radius": int,        // in kilometers (max: 20,037 km)
      "enableAlerts": bool,
      "enablePush": bool
    }
    ```
  - Response:
    ```json
    {
      "status": "success",
      "message": "Location stored successfully"
    }
    ```

##### Earthquakes API
- `GET /api/quakes` - Get earthquakes based on filters
  - Query Parameters:
    - `hours` (int, default: 24) - Time range in hours
    - `minMagnitude` (float, default: 0.0) - Minimum earthquake magnitude
    - `maxMagnitude` (float, default: 10.0) - Maximum earthquake magnitude
    - `radius` (int, default: 100) - Search radius in kilometers
  - Response: Array of earthquake objects
    ```json
    [
      {
        "magnitude": float64,
        "place": string,
        "time": string,      // ISO 8601 format
        "latitude": float64,
        "longitude": float64,
        "depth": float64,    // in kilometers
        "timeAgo": string,   // human-readable time
        "distance": float64  // distance from user in km (if location provided)
      }
    ]
    ```

### Security Notes
- All endpoints use secure session management
- Location data is protected with mutex locks
- Sessions are managed via secure HTTP-only cookies
- API responses include proper CORS headers
- Rate limiting is applied to prevent abuse

## Mobile Support

The application is fully responsive and optimized for mobile devices:
- Touch-friendly navigation
- Responsive layouts
- Optimized performance
- Mobile-first design approach
- Geolocation support
- Push notifications

## Development

### Color Palette

```css
--bittersweet: #ef6461ff;
--earth-yellow: #e4b363ff;
--antiflash-white: #e8e9ebff;
--alabaster: #e0dfd5ff;
--onyx: #313638ff;
```

### Security Considerations

- User locations are stored in memory (in production, use a proper database)
- Geolocation data is only used for alert purposes
- Push notifications require explicit user permission
- All sensitive operations are protected with mutex locks
- Environment variables are used for sensitive configuration

## Deployment (Cloud)

### Leapcell Deployment

App is running at https://quakemeup.leapcell.app/
Leapcell Dashboard: https://leapcell.io/workspace/wsp1940861468352741376/dashboard

3. Configure environment variables in Leapcell:
   - Go to your project settings in Dashboard
   ```
    https://leapcell.io/workspace/wsp1940861468352741376/dashboard
    ```
   - Add the following environment variables:
     ```
     MAPBOX_TOKEN=your_mapbox_token
     ```


## Changelog

### [1.1.2] - 2025-07-03
- Deployed on leapcell (SSO via GitHub) - listening to main branch

### [1.1.0] - 2025-07-03
- Optimized map default zoom level to ~500km range view
- Added favicon and apple-touch-icon support
- Enhanced mobile web app capabilities with site manifest
- Improved initial map view for better local earthquake monitoring
- Updated map zoom levels for better user experience
- Added "Go to" button in latest earthquake overlay for quick navigation
- Enhanced latest earthquake visibility with distinctive blue marker
- Improved user location handling with persistent storage and live updates

### [1.0.0] - 2025-07-02
- Added version management with `.version` file
- Enhanced application initialization with version reading
- Updated footer to display current application version
- Improved template handling with centralized version display
- Added proper error handling for version file reading
- Implemented version display across all pages
- Added new interactive map page using Mapbox GL JS
- Integrated Mapbox token management via environment variables
- Enhanced map visualization with custom earthquake markers
- Added real-time earthquake data overlay on Mapbox map
- Implemented secure token handling for Mapbox integration

### [0.5.11] - 2024-03-19
- Added local copies of partner organization logos
- Created static/images/logos directory for logo storage
- Updated About page to use local logo files
- Improved page load performance with local assets

### [0.5.10] - 2024-03-19
- Added partner organizations section to About page
- Added USGS, NOAA, IRIS, and UNESCO logos
- Enhanced About page with interactive partner logos
- Improved About page layout with responsive grid

### [0.5.9] - 2024-03-19
- Added About page with mission statement and platform goals
- Added route handler for /about endpoint
- Enhanced documentation about Earth science awareness
- Added climate change connection information
- Added weather event integration details
- Added community engagement section

### [0.5.8] - 2024-03-19
- Removed Location Settings section from home page
- Simplified home page interface
- Improved page focus on core features

### [0.5.7] - 2024-03-19
- Refactored earthquake API endpoints to use shared functionality
- Enhanced earthquake data filtering with time and magnitude ranges
- Improved error handling and logging for API requests
- Added support for limiting the number of returned earthquakes
- Optimized USGS API integration

### [0.5.6] - 2024-03-19
- Added new alerts page
- Added latest earthquake display with magnitude focus
- Implemented alert settings with localStorage
- Added push notification support
- Enhanced alert display with animations
- Added daily earthquake tracking
- Added proper error handling and logging

### [0.5.5] - 2024-03-19
- Updated color scheme to new palette
- Enhanced visibility of magnitude indicators
- Improved contrast for better readability
- Updated UI elements with new colors
- Maintained map background colors
- Enhanced tooltip and hover state colors

### [0.5.4] - 2024-03-19
- Improved magnitude indicator visibility
- Updated magnitude colors to use lighter palette
- Enhanced magnitude glow effects
- Added progressive intensity for higher magnitudes
- Implemented alternating colors for better distinction

### [0.5.3] - 2024-03-19
- Fixed world map display issues
- Added dynamic SVG map loading
- Improved map container structure
- Enhanced map styling and responsiveness
- Added proper error handling for map loading
- Improved map integration with markers

### [0.5.2] - 2024-03-19
- Fixed filter refresh functionality
- Enhanced event listeners for all filter controls
- Added debug logging for filter changes
- Improved debounce function reliability
- Fixed API request triggering
- Added proper filter validation before updates

### [0.5.1] - 2024-03-19
- Removed Apply Filters button in favor of automatic updates
- Implemented real-time filter updates with debouncing
- Enhanced filter persistence and loading
- Improved user experience with immediate feedback
- Optimized API calls with 300ms debounce
- Maintained filter state across page reloads

### [0.5.0] - 2024-03-19
- Added real-time earthquake count display
- Enhanced filter functionality with automatic updates
- Added debouncing for filter changes
- Improved page title with earthquake count
- Added loading and error states for data fetching
- Removed manual filter apply button in favor of automatic updates
- Enhanced user feedback during data loading

### [0.4.9] - 2024-03-19
- Added comprehensive filter section to map page
- Implemented radius filter with 10-20,037km range
- Added duration filter with multiple time periods
- Added magnitude range filter with min/max controls
- Added filter persistence using localStorage
- Enhanced API integration with filter parameters
- Improved user interface with consistent styling

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

## License

This project is licensed under the MIT License - see the LICENSE file for details.
