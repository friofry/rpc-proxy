package confighttpserver

import (
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
}

func New(port, providersPath string) *Server {
	return &Server{
		port:          port,
		providersPath: providersPath,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/providers", s.providersHandler)
	http.HandleFunc("/health", s.healthHandler)

	log.Printf("starting config HTTP server on :%s", s.port)
	return http.ListenAndServe(":"+s.port, nil)
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
