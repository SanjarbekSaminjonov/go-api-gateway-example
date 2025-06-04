package main

import (
	"api-gateway/config"
	"api-gateway/gateway"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Mock backend services for demonstration
func startMockService(name string, port string, message string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] Received request: %s %s", name, r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"service": name,
			"message": message,
			"path":    r.URL.Path,
			"method":  r.Method,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	log.Printf("Starting mock service [%s] on port %s", name, port)
	go func() {
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatalf("Failed to start mock service %s: %v", name, err)
		}
	}()
}

func main() {
	log.Println("Starting API Gateway...")

	// Start mock backend services
	// These would be your actual microservices in a real scenario
	startMockService("ServiceA", "8081", "Hello from Service A!")
	startMockService("ServiceB", "8082", "Greetings from Service B!")
	startMockService("ServiceC", "8083", "Welcome to Service C!")

	// Give mock services a moment to start
	time.Sleep(100 * time.Millisecond)

	// Load configuration
	// In a real application, this might come from a file or environment variables
	cfg := &config.Config{
		Routes: []config.Route{
			{PathPrefix: "/service-a", TargetURL: "http://localhost:8081", Name: "ServiceA"},
			{PathPrefix: "/service-b", TargetURL: "http://localhost:8082", Name: "ServiceB"},
			{PathPrefix: "/service-c/specific", TargetURL: "http://localhost:8083", Name: "ServiceCSpecific", StripPrefix: "/service-c/specific"}, // More specific route
			{PathPrefix: "/service-c", TargetURL: "http://localhost:8083", Name: "ServiceCGeneral", StripPrefix: "/service-c"},                    // General route for Service C
		},
		Port: "80",
	}

	// Initialize the gateway router
	gwRouter, err := gateway.NewRouter(cfg.Routes)
	if err != nil {
		log.Fatalf("Failed to initialize gateway router: %v", err)
	}

	// Create the main handler for the API gateway
	// This handler will use the router to decide where to forward requests
	mainHandler := gateway.NewGatewayHandler(gwRouter)

	// Set up the server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mainHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("API Gateway listening on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", cfg.Port, err)
	}

	log.Println("API Gateway shut down.")
}
