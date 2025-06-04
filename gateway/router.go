package gateway

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"api-gateway/config"
)

// Router holds the routing configuration and decides which backend service
// a request should be forwarded to.
type Router struct {
	Routes []config.Route
}

// NewRouter creates a new Router with the given routes.
// It sorts routes by path prefix length in descending order to ensure
// more specific routes are matched first (e.g., "/service/a/b" before "/service/a").
func NewRouter(routes []config.Route) (*Router, error) {
	// Sort routes by the length of PathPrefix in descending order.
	// This ensures that more specific paths are matched before more general ones.
	// For example, "/api/users/specific" should be matched before "/api/users/".
	sortedRoutes := make([]config.Route, len(routes))
	copy(sortedRoutes, routes)
	sort.SliceStable(sortedRoutes, func(i, j int) bool {
		return len(sortedRoutes[i].PathPrefix) > len(sortedRoutes[j].PathPrefix)
	})

	log.Println("Gateway Router: Initialized with the following routes (sorted by specificity):")
	for i, r := range sortedRoutes {
		log.Printf("  %d: Name: %s, PathPrefix: %s, Target: %s, StripPrefix: %s", i+1, r.Name, r.PathPrefix, r.TargetURL, r.StripPrefix)
	}

	return &Router{
		Routes: sortedRoutes,
	}, nil
}

// Route finds the appropriate backend service (Route) for the given HTTP request.
// It iterates through the configured routes and returns the first one that matches
// the request's URL path prefix.
func (rt *Router) Route(r *http.Request) *config.Route {
	requestPath := r.URL.Path
	log.Printf("Router: Attempting to route path: %s", requestPath)

	for _, route := range rt.Routes {
		// Check if the request path starts with the route's PathPrefix
		if strings.HasPrefix(requestPath, route.PathPrefix) {
			log.Printf("Router: Matched route '%s' for path '%s' with prefix '%s'", route.Name, requestPath, route.PathPrefix)
			return &route // Return a pointer to a copy of the route
		}
	}

	log.Printf("Router: No route matched for path: %s", requestPath)
	return nil // No route found
}