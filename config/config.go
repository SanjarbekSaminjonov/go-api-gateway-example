package config

import (
	"log"
)

// Config holds the configuration for the API Gateway.
type Config struct {
	Routes []Route `json:"routes"` // List of routes the gateway will manage
	Port   string  `json:"port"`   // Port on which the gateway will listen
}

// Route defines a single route from a public path prefix to a backend service URL.
type Route struct {
	Name        string `json:"name"`        // A descriptive name for the route (e.g., "UserService", "ProductService")
	PathPrefix  string `json:"pathPrefix"`  // The URL path prefix that this route matches (e.g., "/users/", "/products/")
	TargetURL   string `json:"targetURL"`   // The base URL of the backend service (e.g., "http://localhost:8081", "[http://user-service.internal:80](http://user-service.internal:80)")
	StripPrefix string `json:"stripPrefix"` // Optional: The prefix to strip from the request URL path before forwarding to the target. If empty, PathPrefix is stripped.
}

// LoadConfig would typically load configuration from a file (e.g., JSON, YAML) or environment variables.
// For this basic example, we'll hardcode it in main.go.
func LoadConfig(filePath string) (*Config, error) {
	// In a real application, you would read and parse a config file here.
	// For example, using os.ReadFile and json.Unmarshal.
	log.Printf("Config: Loading configuration (currently hardcoded in main for this example)")
	// Placeholder for actual file loading logic
	return nil, nil // This function is not used in the current basic setup
}
