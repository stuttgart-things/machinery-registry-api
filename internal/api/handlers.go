package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/stuttgart-things/machinery-registry-api/internal/registry"
)

// ClaimListResponse wraps claims for the list endpoint
type ClaimListResponse struct {
	APIVersion string                `json:"apiVersion"`
	Kind       string                `json:"kind"`
	Items      []registry.ClaimEntry `json:"items"`
}

// listClaims returns all claims, optionally filtered by query parameters.
func (s *Server) listClaims(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reg := s.syncer.GetRegistry()
	if reg == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "registry not yet loaded",
		})
		return
	}

	category := r.URL.Query().Get("category")
	template := r.URL.Query().Get("template")
	status := r.URL.Query().Get("status")
	source := r.URL.Query().Get("source")

	items := registry.FilterEntries(reg, category, template, status, source)
	if items == nil {
		items = []registry.ClaimEntry{}
	}

	response := ClaimListResponse{
		APIVersion: "claim-registry.io/v1alpha1",
		Kind:       "ClaimList",
		Items:      items,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getClaim returns a single claim by name.
func (s *Server) getClaim(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reg := s.syncer.GetRegistry()
	if reg == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "registry not yet loaded",
		})
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	entry := registry.FindEntry(reg, name)
	if entry == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "claim not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entry)
}
