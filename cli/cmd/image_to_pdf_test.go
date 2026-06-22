package cmd

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
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

func TestImageToSinglePagePDFAppliesEXIFOrientation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oriented.jpg")
	var img bytes.Buffer
	src := image.NewRGBA(image.Rect(0, 0, 4, 3))
	if err := jpeg.Encode(&img, src, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	data := append([]byte{}, img.Bytes()[:2]...)
	data = append(data, exifOrientationSegment(6)...)
	data = append(data, img.Bytes()[2:]...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write jpeg: %v", err)
	}

	pdf, err := imageToSinglePagePDF(path, 72)
	if err != nil {
		t.Fatalf("imageToSinglePagePDF: %v", err)
	}
	s := string(pdf)
	for _, want := range []string{"/Width 3", "/Height 4", "/MediaBox [0 0 3.0000 4.0000]"} {
		if !strings.Contains(s, want) {
			t.Errorf("PDF did not apply EXIF orientation, missing %q:\n%s", want, s)
		}
	}
}

func TestTIFFOrientation(t *testing.T) {
	data := tiffOrientationPayload(binary.BigEndian, 8)
	if got := tiffOrientation(data); got != 8 {
		t.Fatalf("tiffOrientation = %d, want 8", got)
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

func exifOrientationSegment(orientation uint16) []byte {
	payload := append([]byte("Exif\x00\x00"), tiffOrientationPayload(binary.BigEndian, orientation)...)
	seg := []byte{0xff, 0xe1, 0, 0}
	binary.BigEndian.PutUint16(seg[2:4], uint16(len(payload)+2))
	return append(seg, payload...)
}

func tiffOrientationPayload(order binary.ByteOrder, orientation uint16) []byte {
	data := make([]byte, 8+2+12+4)
	if order == binary.LittleEndian {
		data[0], data[1] = 'I', 'I'
	} else {
		data[0], data[1] = 'M', 'M'
	}
	order.PutUint16(data[2:4], 42)
	order.PutUint32(data[4:8], 8)
	order.PutUint16(data[8:10], 1)
	entry := data[10:22]
	order.PutUint16(entry[0:2], 274)
	order.PutUint16(entry[2:4], 3)
	order.PutUint32(entry[4:8], 1)
	order.PutUint16(entry[8:10], orientation)
	return data
}
