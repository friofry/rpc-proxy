package confighttpserver

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
)

type Provider struct {
	URL        string `json:"url"`
	AuthHeader string `json:"auth_header"`
}

type Server struct {
	port          string
	providersPath string
	server        *http.Server
}

func New(port, providersPath string) *Server {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	s := &Server{
		port:          port,
		providersPath: providersPath,
		server:        srv,
	}

	mux.HandleFunc("/providers", s.providersHandler)
	mux.HandleFunc("/health", s.healthHandler)

	return s
}

func (s *Server) Start() error {
	log.Printf("starting config HTTP server on :%s", s.port)
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(context.Background())
}

func (s *Server) providersHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(s.providersPath)
	if err != nil {
		http.Error(w, "failed to open providers file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, "failed to read providers.json", http.StatusInternalServerError)
		return
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
