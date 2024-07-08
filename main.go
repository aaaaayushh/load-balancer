package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var backendServers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
}

var currBackend uint64

func LoadBalancer(w http.ResponseWriter, r *http.Request) {
	backend := backendServers[atomic.AddUint64(&currBackend, 1)%uint64(len(backendServers))]
	url, err := url.Parse(backend)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)
	fmt.Printf("forwarded request to backend: %s\n", url)
}

func BackendHandler(port string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from backend server on port %s\n", port)
		fmt.Printf("Received request on backend server port %s\n", port)
	}
}

func main() {
	// Start backend servers
	for _, addr := range backendServers {
		go func(addr string) {
			url, _ := url.Parse(addr)
			port := url.Port()
			fmt.Printf("Starting backend server on %s\n", port)
			http.ListenAndServe(":"+port, BackendHandler(port))
		}(addr)
	}

	// start load balancer
	fmt.Println("Starting load balancer on :8080")
	http.HandleFunc("/", LoadBalancer)
	err := http.ListenAndServe(":8080", http.HandlerFunc(LoadBalancer))
	if err != nil {
		panic(err)
	}
}
