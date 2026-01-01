package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

func main() {
	logger.Init("api-gateway")
	logger.Info("Starting API Gateway", "port", 8080)

	// Simple Reverse Proxy Map
	// In production, use a robust router (Chi/Gorilla) and service discovery
	http.HandleFunc("/v1/auth/", proxyRequest("http://localhost:8081"))  // Identity Service
	http.HandleFunc("/v1/orgs", proxyRequest("http://localhost:8081"))   // Identity Service
	http.HandleFunc("/v1/events", proxyRequest("http://localhost:8082")) // Catalog Service
	http.HandleFunc("/v1/trips", proxyRequest("http://localhost:8082"))  // Catalog Service

	http.ListenAndServe(":8080", nil)
}

func proxyRequest(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(url)

		// Optional: Enrich headers here (e.g. inject Request-ID)
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))

		proxy.ServeHTTP(w, r)
	}
}
