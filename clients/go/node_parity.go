package retab

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

type SignatureVerificationError struct {
	Message string
}

func (e SignatureVerificationError) Error() string {
	return e.Message
}

// VerifyEvent validates a Retab webhook HMAC signature and unmarshals the payload.
func VerifyEvent[T any](eventBody []byte, eventSignature string, secret string) (*T, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(eventBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(eventSignature), []byte(expected)) {
		return nil, SignatureVerificationError{Message: "Invalid signature"}
	}
	var event T
	if err := json.Unmarshal(eventBody, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

type Stream[T any] struct {
	decoder *json.Decoder
	closer  io.Closer
}

func (s *Stream[T]) Next() (*T, error) {
	var item T
	if err := s.decoder.Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Stream[T]) Close() error {
	return s.closer.Close()
}

func (c *Client) doStream(ctx context.Context, method string, path string, query url.Values, body any, opts ...RequestOption) (*Stream[Resource], error) {
	req, err := c.newRequest(ctx, method, path, query, body, opts...)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, parseAPIError(resp, responseBody)
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/stream+json") {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("retab: response is not stream JSON")
	}
	return &Stream[Resource]{decoder: json.NewDecoder(resp.Body), closer: resp.Body}, nil
}

type multipartPart struct {
	fieldName string
	filename  string
	mimeType  string
	reader    io.Reader
	value     string
	isFile    bool
}

func (c *Client) doMultipart(ctx context.Context, method string, path string, query url.Values, parts []multipartPart, out any, opts ...RequestOption) error {
	req, err := c.newMultipartRequest(ctx, method, path, query, parts, opts...)
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

func (c *Client) doMultipartStream(ctx context.Context, method string, path string, query url.Values, parts []multipartPart, opts ...RequestOption) (*Stream[Resource], error) {
	req, err := c.newMultipartRequest(ctx, method, path, query, parts, opts...)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, parseAPIError(resp, responseBody)
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/stream+json") {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("retab: response is not stream JSON")
	}
	return &Stream[Resource]{decoder: json.NewDecoder(resp.Body), closer: resp.Body}, nil
}

func (c *Client) newMultipartRequest(ctx context.Context, method string, path string, query url.Values, parts []multipartPart, opts ...RequestOption) (*http.Request, error) {
	options := collectRequestOptions(opts)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		if part.isFile {
			header := make(textproto.MIMEHeader)
			header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(part.fieldName), escapeQuotes(part.filename)))
			if part.mimeType != "" {
				header.Set("Content-Type", part.mimeType)
			}
			partWriter, err := writer.CreatePart(header)
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(partWriter, part.reader); err != nil {
				return nil, err
			}
		} else if err := writer.WriteField(part.fieldName, part.value); err != nil {
			return nil, err
		}
	}
	for key, value := range options.Body {
		stringValue, ok := value.(string)
		if !ok {
			encoded, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			stringValue = string(encoded)
		}
		if err := writer.WriteField(key, stringValue); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	opts = append(opts, WithRequestHeader("Content-Type", writer.FormDataContentType()))
	req, err := c.newRequest(ctx, method, path, query, nil, opts...)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(&body)
	req.ContentLength = int64(body.Len())
	return req, nil
}

func escapeQuotes(value string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(value)
}

type Partition = Resource

type PartitionsService struct {
	client *Client
}

type PartitionCreateRequest struct {
	Document     any    `json:"document"`
	Key          string `json:"key"`
	Instructions string `json:"instructions"`
	Model        string `json:"model"`
	NConsensus   int    `json:"n_consensus,omitempty"`
	BustCache    bool   `json:"bust_cache,omitempty"`
}

func (s *PartitionsService) Create(ctx context.Context, request PartitionCreateRequest, opts ...RequestOption) (*Partition, error) {
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	if request.Key == "" {
		return nil, fmt.Errorf("retab: key is required")
	}
	if request.Instructions == "" {
		return nil, fmt.Errorf("retab: instructions are required")
	}
	if request.Model == "" {
		return nil, fmt.Errorf("retab: model is required")
	}
	if request.NConsensus == 0 {
		request.NConsensus = 1
	}
	var result Partition
	err := s.client.do(ctx, http.MethodPost, "/partitions", nil, request, &result, opts...)
	return &result, err
}

func (s *PartitionsService) Get(ctx context.Context, partitionID string, opts ...RequestOption) (*Partition, error) {
	if partitionID == "" {
		return nil, fmt.Errorf("retab: partitionID is required")
	}
	var result Partition
	err := s.client.do(ctx, http.MethodGet, "/partitions/"+url.PathEscape(partitionID), nil, nil, &result, opts...)
	return &result, err
}

func (s *PartitionsService) List(ctx context.Context, params *ListParams, opts ...RequestOption) (*PaginatedList[Partition], error) {
	query := listQuery(params)
	var result PaginatedList[Partition]
	err := s.client.do(ctx, http.MethodGet, "/partitions", query, nil, &result, opts...)
	return &result, err
}

func (s *PartitionsService) Delete(ctx context.Context, partitionID string, opts ...RequestOption) error {
	if partitionID == "" {
		return fmt.Errorf("retab: partitionID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/partitions/"+url.PathEscape(partitionID), nil, nil, nil, opts...)
}

type EditTemplate = Resource
type FormField = Resource

type EditTemplatesService struct {
	client *Client
}

type EditTemplateCreateRequest struct {
	Name       string      `json:"name"`
	Document   any         `json:"document"`
	FormFields []FormField `json:"form_fields"`
}

type EditTemplateUpdateRequest struct {
	Name       *string     `json:"name,omitempty"`
	FormFields []FormField `json:"form_fields,omitempty"`
}

type EditTemplateFillRequest struct {
	TemplateID   string `json:"template_id"`
	Instructions string `json:"instructions"`
	Model        string `json:"model,omitempty"`
	Color        string `json:"-"`
	BustCache    bool   `json:"bust_cache,omitempty"`
}

type ListEditTemplatesParams struct {
	ListParams
	Name string
}

func (s *EditTemplatesService) Create(ctx context.Context, request EditTemplateCreateRequest, opts ...RequestOption) (*EditTemplate, error) {
	if request.Name == "" {
		return nil, fmt.Errorf("retab: name is required")
	}
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	body := resourceFromJSON(request)
	var result EditTemplate
	err := s.client.do(ctx, http.MethodPost, "/edits/templates", nil, body, &result, opts...)
	return &result, err
}

func (s *EditTemplatesService) Get(ctx context.Context, templateID string, opts ...RequestOption) (*EditTemplate, error) {
	if templateID == "" {
		return nil, fmt.Errorf("retab: templateID is required")
	}
	var result EditTemplate
	err := s.client.do(ctx, http.MethodGet, "/edits/templates/"+url.PathEscape(templateID), nil, nil, &result, opts...)
	return &result, err
}

func (s *EditTemplatesService) List(ctx context.Context, params *ListEditTemplatesParams, opts ...RequestOption) (*PaginatedList[EditTemplate], error) {
	query := listQuery(nil)
	if params != nil {
		applyListParams(query, &params.ListParams)
		addQuery(query, "name", params.Name)
	}
	var result PaginatedList[EditTemplate]
	err := s.client.do(ctx, http.MethodGet, "/edits/templates", query, nil, &result, opts...)
	return &result, err
}

func (s *EditTemplatesService) Update(ctx context.Context, templateID string, request EditTemplateUpdateRequest, opts ...RequestOption) (*EditTemplate, error) {
	if templateID == "" {
		return nil, fmt.Errorf("retab: templateID is required")
	}
	body := resourceFromJSON(request)
	var result EditTemplate
	err := s.client.do(ctx, http.MethodPatch, "/edits/templates/"+url.PathEscape(templateID), nil, body, &result, opts...)
	return &result, err
}

func (s *EditTemplatesService) Delete(ctx context.Context, templateID string, opts ...RequestOption) error {
	if templateID == "" {
		return fmt.Errorf("retab: templateID is required")
	}
	return s.client.do(ctx, http.MethodDelete, "/edits/templates/"+url.PathEscape(templateID), nil, nil, nil, opts...)
}

func (s *EditTemplatesService) Fill(ctx context.Context, request EditTemplateFillRequest, opts ...RequestOption) (*Edit, error) {
	if request.TemplateID == "" {
		return nil, fmt.Errorf("retab: templateID is required")
	}
	if request.Instructions == "" {
		return nil, fmt.Errorf("retab: instructions are required")
	}
	body := resourceFromJSON(request)
	if request.Color != "" {
		body["config"] = Resource{"color": request.Color}
	}
	var result Edit
	err := s.client.do(ctx, http.MethodPost, "/edits/templates/fill", nil, body, &result, opts...)
	return &result, err
}

func (s *ExtractionsService) CreateStream(ctx context.Context, request ExtractionCreateRequest, opts ...RequestOption) (*Stream[Resource], error) {
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	if request.JSONSchema == nil {
		return nil, fmt.Errorf("retab: jsonSchema is required")
	}
	if request.Model == "" {
		return nil, fmt.Errorf("retab: model is required")
	}
	body := resourceFromJSON(request)
	body["stream"] = true
	return s.client.doStream(ctx, http.MethodPost, "/extractions/stream", nil, body, opts...)
}

func (s *DocumentsService) ExtractStream(ctx context.Context, request DocumentExtractRequest, opts ...RequestOption) (*Stream[Resource], error) {
	if request.Document == nil {
		return nil, fmt.Errorf("retab: document is required")
	}
	if request.JSONSchema == nil {
		return nil, fmt.Errorf("retab: jsonSchema is required")
	}
	body := resourceFromJSON(request)
	body["stream"] = true
	return s.client.doStream(ctx, http.MethodPost, "/documents/extractions", nil, body, opts...)
}
