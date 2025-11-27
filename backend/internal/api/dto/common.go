package dto

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ==================== Validation Errors ====================

var (
	ErrInvalidUUID   = errors.New("invalid UUID format")
	ErrRequiredField = errors.New("required field is missing")
)

// ==================== Pagination ====================

type PaginationRequest struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

func (p *PaginationRequest) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 50
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

func ParsePagination(r *http.Request) PaginationRequest {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	p := PaginationRequest{Page: page, PerPage: perPage}
	p.Normalize()
	return p
}

// ==================== ID Extraction ====================

func ParseUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, param)
	if idStr == "" {
		return uuid.Nil, ErrRequiredField
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	return id, nil
}

// ==================== JSON Parsing ====================

func ParseJSON[T any](r *http.Request) (*T, error) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

// ==================== String Validation ====================

func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

func ValidateMaxLength(value string, max int, fieldName string) error {
	if len(value) > max {
		return errors.New(fieldName + " exceeds maximum length")
	}
	return nil
}

// ==================== List Response ====================

type ListMeta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasMore    bool `json:"has_more"`
}

func NewListMeta(page, perPage, total int) ListMeta {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}
	return ListMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
	}
}
