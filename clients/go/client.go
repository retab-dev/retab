package retab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.retab.com/v1"

// Client is the root Retab API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	headers    map[string]string

	Files           *FilesService
	Schemas         *SchemasService
	Extractions     *ExtractionsService
	Partitions      *PartitionsService
	Splits          *SplitsService
	Classifications *ClassificationsService
	Parses          *ParsesService
	Edits           *EditsService
	Jobs            *JobsService
	Workflows       *WorkflowsService
}

// Option customizes a Retab client.
type Option func(*Client)

// WithBaseURL overrides the default Retab API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient overrides the HTTP client used for requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithHeader adds a header to every request.
func WithHeader(key string, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// NewClient creates a Retab client. If apiKey is empty, RETAB_API_KEY is used.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("RETAB_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("retab: api key is required; pass one to NewClient or set RETAB_API_KEY")
	}

	client := &Client{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Minute,
		},
		headers: map[string]string{},
	}
	for _, opt := range opts {
		opt(client)
	}
	client.Files = &FilesService{client: client}
	client.Schemas = &SchemasService{client: client}
	client.Extractions = &ExtractionsService{client: client}
	client.Partitions = &PartitionsService{client: client}
	client.Splits = &SplitsService{client: client}
	client.Classifications = &ClassificationsService{client: client}
	client.Parses = &ParsesService{client: client}
	client.Edits = newEditsService(client)
	client.Jobs = &JobsService{client: client}
	client.Workflows = newWorkflowsService(client)
	return client, nil
}

// RequestOptions mirrors the Node SDK's per-call escape hatch.
type RequestOptions struct {
	Params  url.Values
	Headers map[string]string
	Body    map[string]any
}

// RequestOption customizes one API request.
type RequestOption func(*RequestOptions)

// PreparedRequest is a request descriptor matching the Node SDK prepare_* shape.
type PreparedRequest struct {
	URL     string
	Method  string
	Params  url.Values
	Headers map[string]string
	Body    any
}

// WithRequestParams merges query parameters into one API request.
func WithRequestParams(params url.Values) RequestOption {
	return func(options *RequestOptions) {
		for key, values := range params {
			for _, value := range values {
				options.Params.Add(key, value)
			}
		}
	}
}

// WithRequestHeader sets one header on one API request.
func WithRequestHeader(key string, value string) RequestOption {
	return func(options *RequestOptions) {
		options.Headers[key] = value
	}
}

// WithRequestBody merges JSON object fields into one non-GET request body.
func WithRequestBody(body map[string]any) RequestOption {
	return func(options *RequestOptions) {
		for key, value := range body {
			options.Body[key] = value
		}
	}
}

func collectRequestOptions(opts []RequestOption) RequestOptions {
	options := RequestOptions{
		Params:  url.Values{},
		Headers: map[string]string{},
		Body:    map[string]any{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func mergeQuery(query url.Values, options RequestOptions) url.Values {
	merged := url.Values{}
	for key, values := range query {
		for _, value := range values {
			merged.Add(key, value)
		}
	}
	for key, values := range options.Params {
		merged.Del(key)
		for _, value := range values {
			merged.Add(key, value)
		}
	}
	return merged
}

func mergeBody(body any, options RequestOptions) any {
	if len(options.Body) == 0 {
		return body
	}
	merged := map[string]any{}
	switch typed := body.(type) {
	case nil:
	case map[string]any:
		for key, value := range typed {
			merged[key] = value
		}
	case Resource:
		for key, value := range typed {
			merged[key] = value
		}
	default:
		bodyBytes, err := json.Marshal(typed)
		if err == nil {
			_ = json.Unmarshal(bodyBytes, &merged)
		}
	}
	for key, value := range options.Body {
		merged[key] = value
	}
	return merged
}

func (c *Client) newRequest(ctx context.Context, method string, path string, query url.Values, body any, opts ...RequestOption) (*http.Request, error) {
	options := collectRequestOptions(opts)
	query = mergeQuery(query, options)
	body = mergeBody(body, options)
	if body == nil && method != http.MethodGet {
		body = map[string]any{}
	}
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("retab: invalid base URL: %w", err)
	}
	if baseURL.Path != "" && !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	relativeURL, err := url.Parse(strings.TrimLeft(path, "/"))
	if err != nil {
		return nil, fmt.Errorf("retab: invalid request path: %w", err)
	}
	requestURL := baseURL.ResolveReference(relativeURL)
	if query != nil {
		requestURL.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("retab: encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, method string, path string, query url.Values, body any, out any, opts ...RequestOption) error {
	req, err := c.newRequest(ctx, method, path, query, body, opts...)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp, responseBody)
	}
	if out == nil || len(responseBody) == 0 || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Response is not JSON",
			Body:       string(responseBody),
			RequestID:  resp.Header.Get("x-request-id"),
			Method:     resp.Request.Method,
			URL:        resp.Request.URL.String(),
		}
	}
	if err := json.Unmarshal(responseBody, out); err != nil {
		return fmt.Errorf("retab: decode response: %w", err)
	}
	return nil
}
