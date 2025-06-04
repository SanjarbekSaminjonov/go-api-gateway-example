package gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// GatewayHandler is the main HTTP handler for the API gateway.
// It uses the Router to find the appropriate backend service for an incoming request
// and then proxies the request to that service.
type GatewayHandler struct {
	Router *Router
}

// NewGatewayHandler creates a new GatewayHandler.
func NewGatewayHandler(router *Router) *GatewayHandler {
	return &GatewayHandler{
		Router: router,
	}
}

// ServeHTTP implements the http.Handler interface.
// It routes the request and proxies it to the target service.
func (gh *GatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Gateway: Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Find the target service based on the request path
	route := gh.Router.Route(r)
	if route == nil {
		log.Printf("Gateway: No route found for path: %s", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]any{
			"error":  "Service not found",
			"path":   r.URL.Path,
			"status": http.StatusNotFound,
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	targetURL, err := url.Parse(route.TargetURL)
	if err != nil {
		log.Printf("Gateway: Invalid target URL for route %s: %v", route.Name, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]any{
			"error":  "Internal server error: Invalid target URL",
			"status": http.StatusInternalServerError,
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("Gateway: Routing to %s (%s) for path %s", route.Name, route.TargetURL, r.URL.Path)

	// Create a reverse proxy
	// httputil.NewSingleHostReverseProxy is a convenient way to proxy requests.
	// It takes care of copying headers, handling responses, etc.
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the request before forwarding
	// This is where you can add/remove headers, modify the path, etc.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Call the original director to set up basic proxying (e.g., Host header)
		originalDirector(req)

		// Set X-Forwarded-Host and other relevant headers
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Origin-Host", targetURL.Host)

		// Pass Authorization header from original request to backend service
		if auth := r.Header.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		// Preserve the original request URI if needed, or modify it
		if route.StripPrefix != "" && strings.HasPrefix(req.URL.Path, route.StripPrefix) {
			originalPath := req.URL.Path
			req.URL.Path = strings.TrimPrefix(req.URL.Path, route.StripPrefix)
			if !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path // Ensure path starts with a slash
			}
			log.Printf("Gateway: Stripped prefix '%s', new path for backend: '%s' (original: '%s')", route.StripPrefix, req.URL.Path, originalPath)
		} else if strings.HasPrefix(req.URL.Path, route.PathPrefix) && route.StripPrefix == "" { // Default stripping if StripPrefix is not set
			originalPath := req.URL.Path
			req.URL.Path = strings.TrimPrefix(req.URL.Path, route.PathPrefix)
			if !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path
			}
			log.Printf("Gateway: Default stripping prefix '%s', new path for backend: '%s' (original: '%s')", route.PathPrefix, req.URL.Path, originalPath)
		}

		// Ensure the raw path is also updated if the path is changed.
		// Some servers or frameworks might rely on RawPath.
		req.URL.RawPath = req.URL.EscapedPath()

		log.Printf("Gateway: Forwarding request to backend: %s %s", targetURL, req.URL.String())
	}

	// Custom error handler for the proxy
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Gateway: Proxy error for %s: %v", route.Name, err)
		// It's important to check if headers have already been written.
		// If so, we can't write a new error status code.
		if rw.Header().Get("status") == "" { // A bit of a hack, check if status is already set
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadGateway)
			resp := map[string]any{
				"error":  "Error connecting to backend service",
				"detail": err.Error(),
				"status": http.StatusBadGateway,
			}
			_ = json.NewEncoder(rw).Encode(resp)
		} else {
			log.Printf("Gateway: Headers already written, cannot send new error to client for proxy error.")
			// You could also consider logging the error or sending a generic error response
			// in this case, since we can't send a proper error response.
		}
	}

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
	log.Printf("Gateway: Finished processing request for %s", r.URL.Path)
}
