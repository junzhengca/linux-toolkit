package server

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func (s *Server) registerRoutes(mux *http.ServeMux, uiEnabled bool) {
	if uiEnabled {
		staticFS, err := fs.Sub(staticFiles, "static")
		if err != nil {
			log.Printf("Failed to create static filesystem: %v", err)
		} else {
			fsHandler := http.FileServer(http.FS(staticFS))
			mux.Handle("/", fsHandler)
			mux.Handle("/static/", http.StripPrefix("/static/", fsHandler))
		}
	}

	apiBase := "/api/v1"
	mux.HandleFunc(fmt.Sprintf("%s/health", apiBase), s.healthCheckHandler())
	mux.HandleFunc(fmt.Sprintf("%s/disk", apiBase), s.handleDisk())
	mux.HandleFunc(fmt.Sprintf("%s/disk/", apiBase), s.handleDisk())
	mux.HandleFunc(fmt.Sprintf("%s/cpu", apiBase), s.handleCPU())
	mux.HandleFunc(fmt.Sprintf("%s/cpu/cores", apiBase), s.handleCores())
	mux.HandleFunc(fmt.Sprintf("%s/gpu", apiBase), s.handleGPU())
	mux.HandleFunc(fmt.Sprintf("%s/gpu/", apiBase), s.handleGPU())
	mux.HandleFunc(fmt.Sprintf("%s/battery", apiBase), s.handleBattery())
	mux.HandleFunc(fmt.Sprintf("%s/services", apiBase), s.handleServices())
	mux.HandleFunc(fmt.Sprintf("%s/services/", apiBase), s.handleServices())
	mux.HandleFunc(fmt.Sprintf("%s/summary", apiBase), s.handleSummary())
}
