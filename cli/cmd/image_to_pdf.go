package cmd

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"os"
	"path/filepath"

	// Decoders registered with image.Decode. Stdlib covers png/jpeg/gif; the
	// golang.org/x/image packages add bmp/tiff/webp — all pure Go, so the
	// CGO_ENABLED=0 static build is preserved.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Why this file exists
// --------------------
// `lit` parses PDFs through its bundled PDFium + Tesseract pipeline with no
// external dependencies, but for a *standalone image* (PNG/JPG/…) it shells out
// to ImageMagick to convert the image first — and ImageMagick is not something
// we can bake into the single `retab` binary without ballooning it by 40-160MB.
//
// So instead of bundling ImageMagick, we wrap a standalone image into a minimal
// single-page PDF here, in pure Go, and hand that to `lit`. The image then flows
// through the exact PDFium+Tesseract path we already ship (and already proved
// works fully offline). Net cost: ~0MB, vs. a multi-fold binary blowup.
//
// The wrap is internal to litCLI.Parse/Screenshot; callers still pass the
// original image path, so document_type and anchor kind stay "image".

// imageToSinglePagePDF decodes the image at imgPath and returns a one-page PDF
// embedding it. The page is sized so that rendering at renderDPI reproduces the
// image at its native pixel resolution (best OCR fidelity).
func imageToSinglePagePDF(imgPath string, renderDPI int) ([]byte, error) {
	data, err := os.ReadFile(imgPath)
	if err != nil {
		return nil, fmt.Errorf("open image %s: %w", filepath.Base(imgPath), err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", filepath.Base(imgPath), err)
	}
	img = applyImageOrientation(img, imageOrientation(data))
	return encodeImagePDF(img, renderDPI)
}

func imageOrientation(data []byte) int {
	if orientation := jpegEXIFOrientation(data); orientation != 0 {
		return orientation
	}
	if orientation := tiffOrientation(data); orientation != 0 {
		return orientation
	}
	return 1
}

func jpegEXIFOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return 0
	}
	for pos := 2; pos+4 <= len(data); {
		if data[pos] != 0xff {
			return 0
		}
		for pos < len(data) && data[pos] == 0xff {
			pos++
		}
		if pos >= len(data) {
			return 0
		}
		marker := data[pos]
		pos++
		if marker == 0xd9 || marker == 0xda {
			return 0
		}
		if pos+2 > len(data) {
			return 0
		}
		segLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))
		pos += 2
		if segLen < 2 || pos+segLen-2 > len(data) {
			return 0
		}
		seg := data[pos : pos+segLen-2]
		if marker == 0xe1 && bytes.HasPrefix(seg, []byte("Exif\x00\x00")) {
			return tiffOrientation(seg[6:])
		}
		pos += segLen - 2
	}
	return 0
}

func tiffOrientation(data []byte) int {
	if len(data) < 8 {
		return 0
	}
	var order binary.ByteOrder
	switch {
	case data[0] == 'I' && data[1] == 'I':
		order = binary.LittleEndian
	case data[0] == 'M' && data[1] == 'M':
		order = binary.BigEndian
	default:
		return 0
	}
	if order.Uint16(data[2:4]) != 42 {
		return 0
	}
	ifdOffset := int(order.Uint32(data[4:8]))
	if ifdOffset < 0 || ifdOffset+2 > len(data) {
		return 0
	}
	count := int(order.Uint16(data[ifdOffset : ifdOffset+2]))
	pos := ifdOffset + 2
	for i := 0; i < count && pos+12 <= len(data); i++ {
		entry := data[pos : pos+12]
		tag := order.Uint16(entry[0:2])
		typ := order.Uint16(entry[2:4])
		n := order.Uint32(entry[4:8])
		if tag == 274 && typ == 3 && n >= 1 {
			orientation := int(order.Uint16(entry[8:10]))
			if orientation >= 1 && orientation <= 8 {
				return orientation
			}
			return 0
		}
		pos += 12
	}
	return 0
}

func applyImageOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return flipImageHorizontal(img)
	case 3:
		return rotateImage180(img)
	case 4:
		return flipImageVertical(img)
	case 5:
		return transposeImage(img)
	case 6:
		return rotateImage90CW(img)
	case 7:
		return transverseImage(img)
	case 8:
		return rotateImage270CW(img)
	default:
		return img
	}
}

// flattenImageToRGB returns img as packed 8-bit DeviceRGB samples (top row
// first, PDF sample order), compositing any transparency over a white
// background. RGBA() returns alpha-PREMULTIPLIED 16-bit channels, so
// out = premult + (0xffff-a) leaves opaque pixels unchanged while flattening
// transparent/semi-transparent pixels to white instead of black — important
// because this PDF feeds OCR. The premultiplied invariant (channel <= a)
// guarantees out stays within 0xffff.
func flattenImageToRGB(img image.Image) []byte {
	bnd := img.Bounds()
	rgb := make([]byte, 0, bnd.Dx()*bnd.Dy()*3)
	for y := bnd.Min.Y; y < bnd.Max.Y; y++ {
		for x := bnd.Min.X; x < bnd.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			inv := 0xffff - a
			r += inv
			g += inv
			b += inv
			rgb = append(rgb, byte(r>>8), byte(g>>8), byte(b>>8))
		}
	}
	return rgb
}

func rotateImage90CW(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(h-1-y, x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

func rotateImage180(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(w-1-x, h-1-y, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

func rotateImage270CW(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(y, w-1-x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

func flipImageHorizontal(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(w-1-x, y, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

func flipImageVertical(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(x, h-1-y, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

func transposeImage(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(y, x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

func transverseImage(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.Set(h-1-y, w-1-x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

// encodeImagePDF builds a single-page PDF that draws img full-bleed. The image
// is embedded as an 8-bit DeviceRGB XObject with FlateDecode, which any decoded
// image format can produce uniformly.
func encodeImagePDF(img image.Image, renderDPI int) ([]byte, error) {
	bnd := img.Bounds()
	w, h := bnd.Dx(), bnd.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("image has zero dimension (%dx%d)", w, h)
	}

	// Flatten to RGB (8 bits/component), top row first — PDF image sample order.
	rgb := flattenImageToRGB(img)

	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	if _, err := zw.Write(rgb); err != nil {
		return nil, fmt.Errorf("compress image stream: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("compress image stream: %w", err)
	}
	comp := zbuf.Bytes()

	if renderDPI <= 0 {
		renderDPI = 72
	}
	// Page dimensions in points (1pt = 1/72in). Sizing the page at
	// w*72/dpi means a render at `dpi` yields back exactly w pixels.
	pw := float64(w) * 72.0 / float64(renderDPI)
	ph := float64(h) * 72.0 / float64(renderDPI)

	var buf bytes.Buffer
	offsets := make([]int, 0, 5)
	addObj := func(body string) {
		offsets = append(offsets, buf.Len())
		buf.WriteString(body)
	}

	buf.WriteString("%PDF-1.7\n")
	addObj("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	addObj("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	addObj(fmt.Sprintf(
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.4f %.4f] "+
			"/Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >>\nendobj\n",
		pw, ph,
	))
	content := fmt.Sprintf("q %.4f 0 0 %.4f 0 0 cm /Im0 Do Q", pw, ph)
	addObj(fmt.Sprintf(
		"4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n",
		len(content), content,
	))

	// Object 5 (image) is written by hand because its stream is binary.
	offsets = append(offsets, buf.Len())
	buf.WriteString(fmt.Sprintf(
		"5 0 obj\n<< /Type /XObject /Subtype /Image /Width %d /Height %d "+
			"/ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n",
		w, h, len(comp),
	))
	buf.Write(comp)
	buf.WriteString("\nendstream\nendobj\n")

	const nObj = 5
	xrefStart := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", nObj+1))
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}
	buf.WriteString(fmt.Sprintf(
		"trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		nObj+1, xrefStart,
	))

	return buf.Bytes(), nil
}

// litInputPath resolves the path to hand `lit`. PDFs (and anything not an image)
// pass through unchanged. Standalone images are wrapped into a temporary
// single-page PDF so they OCR through PDFium+Tesseract without ImageMagick. The
// returned cleanup removes any temp file and must always be called.
func litInputPath(path string, dpi int) (string, func(), error) {
	noop := func() {}
	if detectKind(path) != kindImage {
		return path, noop, nil
	}

	pdf, err := imageToSinglePagePDF(path, dpi)
	if err != nil {
		return "", noop, err
	}
	tmp, err := os.CreateTemp("", "retab-img-*.pdf")
	if err != nil {
		return "", noop, fmt.Errorf("create temp pdf: %w", err)
	}
	name := tmp.Name()
	if _, err := tmp.Write(pdf); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return "", noop, fmt.Errorf("write temp pdf: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return "", noop, fmt.Errorf("write temp pdf: %w", err)
	}
	return name, func() { _ = os.Remove(name) }, nil
}
