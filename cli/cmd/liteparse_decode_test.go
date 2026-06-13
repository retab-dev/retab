package cmd

import (
	"encoding/json"
	"testing"
)

// litV2SampleJSON mirrors the real `lit parse --format json` output of the
// pinned bundle (lit-v2.0.x), which emits camelCase keys: `textItems`,
// `fontName`, `fontSize`. This is the exact shape captured from `lit` v2.0.0
// on a real PDF. The decoder must populate per-item bounding boxes from it so
// that `files parse --bbox` and `files grep --bbox` actually carry positions.
const litV2SampleJSON = `{
  "pages": [
    {
      "page": 1,
      "width": 842.0,
      "height": 595.0,
      "text": "Commande de Transport",
      "textItems": [
        {
          "text": "Commande de Transport",
          "x": 290.08,
          "y": 14.85,
          "width": 240.44,
          "height": 12.53,
          "fontName": "TimesNewRomanPS-BoldMT",
          "fontSize": 14.04,
          "confidence": 1.0
        },
        {
          "text": "n 2604399123",
          "x": 540.0,
          "y": 14.85,
          "width": 90.0,
          "height": 12.53,
          "fontName": "TimesNewRomanPSMT",
          "fontSize": 14.04,
          "confidence": 0.91
        }
      ]
    }
  ]
}`

// TestConvertLiteParseDecodesCamelCaseItems pins the decode contract against the
// pinned lit bundle's actual JSON. A snake_case struct tag silently drops every
// positioned item, breaking `--bbox` and leaving Source mislabeled as
// pdf_text_layer for OCR'd output.
func TestConvertLiteParseDecodesCamelCaseItems(t *testing.T) {
	var raw liteParseJSON
	if err := json.Unmarshal([]byte(litV2SampleJSON), &raw); err != nil {
		t.Fatalf("unmarshal lit json: %v", err)
	}
	result := convertLiteParse(&raw)

	if len(result.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(result.Pages))
	}
	page := result.Pages[0]
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 positioned items decoded, got %d (camelCase textItems not decoded)", len(page.Items))
	}

	first := page.Items[0]
	if first.X == 0 || first.Width == 0 {
		t.Errorf("first item geometry not decoded: x=%v width=%v", first.X, first.Width)
	}
	if first.FontName != "TimesNewRomanPS-BoldMT" {
		t.Errorf("fontName not decoded: got %q", first.FontName)
	}
	if first.FontSize == 0 {
		t.Errorf("fontSize not decoded: got %v", first.FontSize)
	}

	// An item with confidence < 1.0 means OCR was involved; Source must reflect it.
	if result.Source != "ocr" {
		t.Errorf("expected Source=ocr when an item carries OCR confidence, got %q", result.Source)
	}
}
