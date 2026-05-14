package retab

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const retabStorageHost = "storage.retab.com"

// NewMIMEDataFromRetabID builds the same storage URL shape used by the Node SDK.
func NewMIMEDataFromRetabID(organizationID string, fileID string, extension string, filename string) MIMEData {
	extension = strings.TrimPrefix(extension, ".")
	storageName := fileID
	if extension != "" {
		storageName += "." + extension
	}
	return MIMEData{
		Filename: filename,
		URL:      "https://" + retabStorageHost + "/" + organizationID + "/" + storageName,
	}
}

// ID returns the Retab file id encoded in a storage URL, when present.
func (m MIMEData) ID() string {
	fileID := retabStorageFileIDFromURL(m.URL)
	if fileID == "" {
		return ""
	}
	return fileID
}

// InferMIMEData accepts the same practical inputs as the Node SDK: local path,
// HTTPS URL, data URI, base64 string, bytes, or reader.
func InferMIMEData(input any) (MIMEData, error) {
	switch value := input.(type) {
	case MIMEData:
		return value, nil
	case string:
		if strings.HasPrefix(value, "https://") {
			return passthroughHTTPSMIMEData(value), nil
		}
		if strings.HasPrefix(value, "data:") {
			decoded, _, err := dataURIBytes(MIMEData{URL: value})
			if err != nil {
				return MIMEData{}, err
			}
			detectedMimeType, ok := detectKnownContentType(decoded)
			if !ok {
				return MIMEData{}, fmt.Errorf("retab: unable to determine file type")
			}
			filename := "uploaded_file" + extensionFromMIMEType(detectedMimeType)
			return MIMEData{Filename: filename, URL: value}, nil
		}
		if stat, err := os.Stat(value); err == nil && !stat.IsDir() {
			return inferMIMEDataFromFile(value)
		}
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return MIMEData{}, fmt.Errorf("retab: unsupported MIME input string")
		}
		return inferMIMEDataFromBytes(decoded, "")
	case []byte:
		return inferMIMEDataFromBytes(value, "")
	case io.Reader:
		data, err := io.ReadAll(value)
		if err != nil {
			return MIMEData{}, err
		}
		return inferMIMEDataFromBytes(data, "")
	default:
		return MIMEData{}, fmt.Errorf("retab: unsupported MIME input type %T", input)
	}
}

func retabStorageFileIDFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host != retabStorageHost || parsed.RawQuery != "" || parsed.Fragment != "" {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" {
		return ""
	}
	name := parts[1]
	index := strings.LastIndex(name, ".")
	if index <= 0 || index == len(name)-1 {
		return ""
	}
	return name[:index]
}

func passthroughHTTPSMIMEData(rawURL string) MIMEData {
	pathPart := strings.Split(rawURL, "?")[0]
	filename := path.Base(pathPart)
	if filename == "." || filename == "/" || filename == "" {
		filename = "remote_file"
	}
	return MIMEData{Filename: filename, URL: rawURL}
}

func inferMIMEDataFromFile(filePath string) (MIMEData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return MIMEData{}, err
	}
	filename := filepath.Base(filePath)
	mimeType, ok := detectKnownContentType(data)
	if !ok {
		// Content-sniffing failed (commonly: CSV, Excel, .eml, plain
		// text without a clear magic number). Fall back to the file's
		// extension — same strategy contentTypeForFilename uses on the
		// upload path. Strip any charset/param suffix appended by the
		// stdlib (`text/plain; charset=utf-8` → `text/plain`) so the
		// resulting data URI stays clean.
		if extType := mime.TypeByExtension(filepath.Ext(filename)); extType != "" {
			mimeType = strings.TrimSpace(strings.SplitN(extType, ";", 2)[0])
		}
	}
	if mimeType == "" {
		return MIMEData{}, fmt.Errorf("retab: unable to determine file type for %s", filename)
	}
	return MIMEData{
		Filename: filename,
		URL:      "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data),
	}, nil
}

func inferMIMEDataFromBytes(data []byte, filename string) (MIMEData, error) {
	mimeType, ok := detectKnownContentType(data)
	if !ok {
		return MIMEData{}, fmt.Errorf("retab: unable to determine file type")
	}
	if filename == "" {
		filename = "uploaded_file" + extensionFromMIMEType(mimeType)
	}
	return MIMEData{
		Filename: filename,
		URL:      "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data),
	}, nil
}

// detectKnownContentType identifies a MIME type from a byte prefix.
// First checks the explicit magic numbers we care about (fast, reliable
// for the formats most users feed in), then falls back to net/http's
// general-purpose sniffer. We accept any non-empty result that isn't
// `application/octet-stream` (the sniffer's "I don't know" reply) —
// broad enough to cover text/plain, text/csv, text/html, RFC822
// emails, Excel/Word zips, etc. The previous whitelist of just 4
// formats was the root cause of `retab: unable to determine file type`
// on every input the marketing copy promised to support beyond images
// and PDFs.
func detectKnownContentType(data []byte) (string, bool) {
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return "application/pdf", true
	}
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		return "image/png", true
	}
	if len(data) >= 3 && bytes.Equal(data[:3], []byte{0xff, 0xd8, 0xff}) {
		return "image/jpeg", true
	}
	if len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a") {
		return "image/gif", true
	}
	sniff := data
	if len(sniff) > 512 {
		sniff = sniff[:512]
	}
	mimeType := http.DetectContentType(sniff)
	// http.DetectContentType returns `application/octet-stream` when it
	// can't classify the bytes — that's the only result we reject here.
	// Strip any charset/parameter suffix for a clean data-URI value.
	if mimeType == "" || mimeType == "application/octet-stream" {
		return "", false
	}
	if i := strings.Index(mimeType, ";"); i > 0 {
		mimeType = strings.TrimSpace(mimeType[:i])
	}
	return mimeType, true
}

func mimeTypeFromDataURI(dataURI string) string {
	header := strings.SplitN(dataURI, ",", 2)[0]
	header = strings.TrimPrefix(header, "data:")
	if index := strings.Index(header, ";"); index >= 0 {
		return header[:index]
	}
	return header
}

func extensionFromMIMEType(mimeType string) string {
	extensions, _ := mime.ExtensionsByType(mimeType)
	if len(extensions) > 0 {
		return extensions[0]
	}
	return ""
}

func dataURIBytes(mimeData MIMEData) ([]byte, string, error) {
	if !strings.HasPrefix(mimeData.URL, "data:") {
		return nil, "", fmt.Errorf("retab: MIMEData URL is not a data URI")
	}
	parts := strings.SplitN(mimeData.URL, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("retab: invalid data URI")
	}
	mimeType := mimeTypeFromDataURI(parts[0])
	if strings.Contains(parts[0], ";base64") {
		data, err := base64.StdEncoding.DecodeString(parts[1])
		return data, mimeType, err
	}
	data, err := url.QueryUnescape(parts[1])
	return []byte(data), mimeType, err
}

func mimeDataReader(mimeData MIMEData) (io.Reader, string, error) {
	if mimeData.Content != "" {
		data, err := base64.StdEncoding.DecodeString(mimeData.Content)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewReader(data), mimeData.MIMEType, nil
	}
	data, mimeType, err := dataURIBytes(mimeData)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(data), mimeType, nil
}
