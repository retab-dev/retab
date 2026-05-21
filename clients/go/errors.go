package retab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError is returned for non-2xx API responses.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]any
	Body       string
	RequestID  string
	Method     string
	URL        string
}

func (e *APIError) Error() string {
	return e.String()
}

func (e *APIError) String() string {
	lines := []string{fmt.Sprintf("%d — %s", e.StatusCode, e.Message)}
	if e.Method != "" && e.URL != "" {
		lines = append(lines, "  URL:        "+e.Method+" "+e.URL)
	}
	if e.RequestID != "" {
		lines = append(lines, "  Request-ID: "+e.RequestID)
	}
	if e.Code != "" {
		lines = append(lines, "  Code:       "+e.Code)
	}
	if e.Details != nil {
		details, _ := json.Marshal(e.Details)
		lines = append(lines, "  Details:    "+string(details))
	}
	if e.Body != "" {
		body := e.Body
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		lines = append(lines, "  Body:       "+body)
	}
	return strings.Join(lines, "\n")
}

func ParseAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    fmt.Sprintf("Request failed (%d)", resp.StatusCode),
		Body:       string(body),
		RequestID:  resp.Header.Get("x-request-id"),
		Method:     resp.Request.Method,
		URL:        resp.Request.URL.String(),
	}

	var parsed struct {
		Detail  any            `json:"detail"`
		Message string         `json:"message"`
		Code    string         `json:"code"`
		Details map[string]any `json:"details"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		if len(body) > 0 {
			apiErr.Message = string(body)
		}
		return apiErr
	}

	switch detail := parsed.Detail.(type) {
	case string:
		apiErr.Message = detail
	case map[string]any:
		if code, ok := detail["code"].(string); ok {
			apiErr.Code = code
		}
		if message, ok := detail["message"].(string); ok {
			apiErr.Message = message
		}
		if details, ok := detail["details"].(map[string]any); ok {
			apiErr.Details = details
		}
	}

	// Flat envelope fallback: FastAPI request-validation failures (every
	// 422) come back from main_server as {"status_code","message","data"}
	// with no "detail" key at all. Without this branch the real validation
	// message degrades to the generic "Request failed (NNN)". Only consulted
	// when "detail" carried nothing, so the nested shape always wins.
	if parsed.Detail == nil {
		if parsed.Message != "" {
			apiErr.Message = parsed.Message
		}
		if parsed.Code != "" {
			apiErr.Code = parsed.Code
		}
		if parsed.Details != nil {
			apiErr.Details = parsed.Details
		}
	}
	return apiErr
}
