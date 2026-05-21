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

	// tokenProvider, when set, returns a Bearer token at request time.
	// Takes precedence over apiKey: requests use the `Authorization` header
	// and `Api-Key` is omitted. Callable on every request so callers can
	// implement transparent refresh without rebuilding the Client.
	tokenProvider func(context.Context) (string, error)

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

// WithBearerToken authenticates as the given Bearer token. The Client will
// send `Authorization: Bearer <token>` on every request and omit `Api-Key`.
// Use this for OAuth-issued access tokens (e.g. from `retab auth login`).
//
// For tokens that need to be refreshed transparently, use
// WithBearerTokenProvider instead — it's invoked per request so the caller
// can return a fresh token whenever the cached one is near expiry.
func WithBearerToken(token string) Option {
	return func(c *Client) {
		t := token
		c.tokenProvider = func(context.Context) (string, error) { return t, nil }
	}
}

// WithBearerTokenProvider authenticates with a Bearer token that's resolved
// at request time. The provider is called once per outgoing request; if it
// returns an error the request fails before any network IO. This is the
// right hook for OAuth refresh: cache the access token + refresh it when
// `time.Now()` crosses `expires_at - leeway`.
func WithBearerTokenProvider(provider func(context.Context) (string, error)) Option {
	return func(c *Client) {
		c.tokenProvider = provider
	}
}

// NewClient creates a Retab client. By default it authenticates with an API
// key — pass one in, or set RETAB_API_KEY. Pass an empty string AND
// WithBearerToken / WithBearerTokenProvider to authenticate with an OAuth
// access token instead. Exactly one auth mode must end up configured.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("RETAB_API_KEY")
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
	if client.apiKey == "" && client.tokenProvider == nil {
		return nil, fmt.Errorf("retab: no credentials; pass an API key, set RETAB_API_KEY, or use WithBearerToken/WithBearerTokenProvider")
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
	existingQuery := requestURL.Query()
	for key, values := range query {
		existingQuery.Del(key)
		for _, value := range values {
			existingQuery.Add(key, value)
		}
	}
	requestURL.RawQuery = existingQuery.Encode()

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
	// Auth: Bearer token wins over API key. Only one header is sent so the
	// server's auth resolution sees a single unambiguous credential.
	if c.tokenProvider != nil {
		token, err := c.tokenProvider(ctx)
		if err != nil {
			return nil, fmt.Errorf("retab: resolve bearer token: %w", err)
		}
		if token == "" {
			return nil, fmt.Errorf("retab: bearer token provider returned empty token")
		}
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("Api-Key", c.apiKey)
	}
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
		return ParseAPIError(resp, responseBody)
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
