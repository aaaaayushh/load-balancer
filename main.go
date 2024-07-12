package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
	server       *http.Server
}

func (s *Server) SetAlive(alive bool) {
	s.mux.Lock()
	s.Alive = alive
	s.mux.Unlock()
}

func (s *Server) IsAlive() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.Alive
}

func (s *Server) Start(id int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from backend server %d\n", id)
	})

	s.server = &http.Server{
		Addr:    ":" + s.URL.Port(),
		Handler: mux,
	}

	go func() {
		fmt.Printf("Starting backend server %d on %s\n", id, s.URL.Port())
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Server %d unexpectedly stopped: %v\n", id, err)
		}
	}()

	s.SetAlive(true)
}

func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
		s.SetAlive(false)
		fmt.Printf("Server %s has been stopped\n", s.URL)
	}
}

type RoundRobinServerPool struct {
	servers []*Server
	current uint64
}

func (s *RoundRobinServerPool) AddServer(server *Server) {
	s.servers = append(s.servers, server)
}

func (s *RoundRobinServerPool) GetNextServer() *Server {
	l := uint64(len(s.servers))
	for i := uint64(0); i < l; i++ {
		idx := atomic.AddUint64(&s.current, 1) % l
		if s.servers[idx].IsAlive() {
			return s.servers[idx]
		}
	}
	return nil
}

func (s *RoundRobinServerPool) HealthCheck() {
	for _, server := range s.servers {
		status := "up"
		alive := isServerAlive(server.URL)
		server.SetAlive(alive)
		if !alive {
			status = "down"
		}
		fmt.Printf("%s [%s]\n", server.URL, status)
	}
}

func isServerAlive(u *url.URL) bool {
	resp, err := http.Head(u.String())
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func LoadBalancer(w http.ResponseWriter, r *http.Request) {
	peer := RoundRobinServerPoolInstance.GetNextServer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func simulateServerDowntime() {
	for {
		time.Sleep(time.Second * 60) // Wait for 10 seconds between simulations
		serverIndex := rand.Intn(len(RoundRobinServerPoolInstance.servers))
		server := RoundRobinServerPoolInstance.servers[serverIndex]

		if server.IsAlive() {
			server.Stop()

			// Simulate server coming back up after 60 seconds
			go func(s *Server, id int) {
				time.Sleep(time.Second * 60)
				s.Start(id)
			}(server, serverIndex)
		}
	}
}

var RoundRobinServerPoolInstance RoundRobinServerPool

func main() {
	servers := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
		"http://localhost:8084",
	}

	for i, s := range servers {
		u, _ := url.Parse(s)
		proxy := httputil.NewSingleHostReverseProxy(u)
		server := &Server{
			URL:          u,
			ReverseProxy: proxy,
		}
		RoundRobinServerPoolInstance.AddServer(server)
		server.Start(i)
	}

	// Start the downtime simulation
	go simulateServerDowntime()

	// Start health check
	go func() {
		for {
			time.Sleep(time.Second * 5)
			RoundRobinServerPoolInstance.HealthCheck()
		}
	}()

	// Start load balancer
	fmt.Println("Starting load balancer on :8080")
	http.HandleFunc("/", LoadBalancer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
