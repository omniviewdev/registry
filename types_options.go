package registry

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListOptions configures pagination, sorting, and filtering for list endpoints.
type ListOptions struct {
	Page           int
	PerPage        int
	OrderField     string
	OrderDirection string
	Search         string
	Category       string
	Featured       bool
}

func (o *ListOptions) buildQuery() string {
	if o == nil {
		return ""
	}
	q := url.Values{}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(o.PerPage))
	}
	if o.OrderField != "" {
		q.Set("order_field", o.OrderField)
	}
	if o.OrderDirection != "" {
		q.Set("order_direction", o.OrderDirection)
	}
	if o.Search != "" {
		q.Set("search", o.Search)
	}
	if o.Category != "" {
		q.Set("category", o.Category)
	}
	if o.Featured {
		q.Set("featured", "true")
	}
	encoded := q.Encode()
	if encoded == "" {
		return ""
	}
	return fmt.Sprintf("?%s", encoded)
}

// Pagination holds pagination metadata from API responses.
type Pagination struct {
	Page       int32 `json:"page"`
	PerPage    int32 `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int32 `json:"total_pages"`
}

// ListResult is a generic paginated response.
type ListResult[T any] struct {
	Items      []T         `json:"items"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// apiResponse is the envelope used by the API.
type apiResponse struct {
	Success    bool            `json:"success"`
	Data       json.RawMessage `json:"data,omitempty"`
	Message    string          `json:"message,omitempty"`
	Pagination *Pagination     `json:"pagination,omitempty"`
}

// HealthStatus represents the API health check response.
type HealthStatus struct {
	Status string `json:"status"`
}
