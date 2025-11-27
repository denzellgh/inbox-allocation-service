package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIResponse is the standard wrapper for all successful responses
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Meta contains pagination and additional metadata for polling
type Meta struct {
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	Total      int    `json:"total,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
}

// PaginationParams holds pagination input parameters
type PaginationParams struct {
	Page    int
	PerPage int
	Cursor  string
}

// DefaultPagination returns default pagination values
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 50,
	}
}

// MaxPerPage is the maximum items per page (requirement: 100)
const MaxPerPage = 100

// JSON sends a JSON response with the given status code
func JSON(w http.ResponseWriter, status int, data interface{}) {
	response := APIResponse{
		Success:   status >= 200 && status < 300,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
	writeJSON(w, status, response)
}

// JSONWithMeta sends a JSON response with pagination metadata
func JSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	response := APIResponse{
		Success:   status >= 200 && status < 300,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC(),
	}
	writeJSON(w, status, response)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Created sends a 201 Created response with data
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// OK sends a 200 OK response with data
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Log error but can't do much at this point
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
