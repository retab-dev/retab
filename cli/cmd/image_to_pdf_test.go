package cmd

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestPNG writes a w x h solid-color PNG and returns its path.
func writeTestPNG(t *testing.T, dir string, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 40, B: 40, A: 255})
		}
	}
	path := filepath.Join(dir, "pic.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEncodeImagePDFStructure(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	pdf, err := encodeImagePDF(img, 150)
	if err != nil {
		t.Fatalf("encodeImagePDF: %v", err)
	}
	s := string(pdf)
	if !strings.HasPrefix(s, "%PDF-") {
		t.Errorf("missing %%PDF header: %.16q", s)
	}
	if !strings.Contains(s, "%%EOF") {
		t.Error("missing EOF trailer")
	}
	for _, want := range []string{
		"/Type /Catalog", "/Type /Pages", "/Type /Page",
		"/Subtype /Image", "/Width 4", "/Height 3",
		"/ColorSpace /DeviceRGB", "/Filter /FlateDecode",
		"startxref",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("PDF missing %q", want)
		}
	}
	// Page sized so a render at 150 DPI reproduces native px: 4*72/150 = 1.92.
	if !strings.Contains(s, "/MediaBox [0 0 1.9200 1.4400]") {
		t.Errorf("unexpected MediaBox; got:\n%s", s)
	}
}

func TestEncodeImagePDFZeroDimension(t *testing.T) {
	if _, err := encodeImagePDF(image.NewRGBA(image.Rect(0, 0, 0, 0)), 72); err == nil {
		t.Fatal("expected error for zero-dimension image")
	}
}

func TestLitInputPathWrapsImage(t *testing.T) {
	dir := t.TempDir()
	pngPath := writeTestPNG(t, dir, 8, 6)

	resolved, cleanup, err := litInputPath(pngPath, 150)
	if err != nil {
		t.Fatalf("litInputPath: %v", err)
	}
	if resolved == pngPath {
		t.Fatal("image input should be rewritten to a temp PDF, not passed through")
	}
	if !strings.HasSuffix(resolved, ".pdf") {
		t.Errorf("wrapped path should end in .pdf: %q", resolved)
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		t.Fatalf("temp pdf not readable: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("temp file is not a PDF")
	}
	// cleanup removes the temp file.
	cleanup()
	if _, err := os.Stat(resolved); !os.IsNotExist(err) {
		t.Errorf("cleanup did not remove temp pdf (err=%v)", err)
	}
}

func TestLitInputPathPassesThroughPDF(t *testing.T) {
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 real"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolved, cleanup, err := litInputPath(pdfPath, 150)
	if err != nil {
		t.Fatalf("litInputPath: %v", err)
	}
	defer cleanup()
	if resolved != pdfPath {
		t.Errorf("PDF should pass through unchanged: got %q want %q", resolved, pdfPath)
	}
	// Passing through must not delete the original on cleanup.
	cleanup()
	if !fileExists(pdfPath) {
		t.Error("cleanup must not remove a passed-through PDF")
	}
}
