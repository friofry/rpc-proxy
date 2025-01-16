package confighttpserver

import (
	"encoding/json"
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
	defaultPath   string
	referencePath string
	outputPath    string
}

func New(port, defaultPath, referencePath, outputPath string) *Server {
	return &Server{
		port:          port,
		defaultPath:   defaultPath,
		referencePath: referencePath,
		outputPath:    outputPath,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/providers", s.providersHandler)
	http.HandleFunc("/health", s.healthHandler)

	log.Printf("starting config HTTP server on :%s", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) providersHandler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(s.outputPath)
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

func UpdateProviders(defaultPath, referencePath, outputPath string) error {
	defaultData, err := os.ReadFile(defaultPath)
	if err != nil {
		return err
	}

	var providers []Provider
	if err := json.Unmarshal(defaultData, &providers); err != nil {
		return err
	}

	if len(providers) < 2 {
		return nil
	}

	selected := providers[:2]
	outData, err := json.MarshalIndent(selected, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, outData, 0644)
}
