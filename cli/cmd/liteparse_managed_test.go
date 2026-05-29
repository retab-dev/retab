package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// makeTarGzBundle builds an in-memory gzip'd tar containing the given members
// (name -> contents) plus some decoy files that extraction must ignore.
func makeTarGzBundle(t *testing.T, members map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	all := map[string]string{"LICENSE": "apache", "BUNDLE.txt": "notes", "sub/dir/lit": "DECOY"}
	for k, v := range members {
		all[k] = v
	}
	for name, body := range all {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func makeZipBundle(t *testing.T, members map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	all := map[string]string{"LICENSE": "apache", "BUNDLE.txt": "notes"}
	for k, v := range members {
		all[k] = v
	}
	for name, body := range all {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sha256hex(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// ocrBundle is a representative full bundle: binary + pdfium + OCR data.
func ocrTarGz(t *testing.T) []byte {
	return makeTarGzBundle(t, map[string]string{
		"lit":             "#!/bin/sh\necho lit",
		"libpdfium.so":    "PDFIUMBYTES",
		"eng.traineddata": "OCRDATA",
	})
}

func TestStageBundleTarGzExtractsAllMembers(t *testing.T) {
	data := ocrTarGz(t)
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "eng.traineddata"}

	dest := filepath.Join(t.TempDir(), "lit-v9.9.9", "linux-amd64")
	if err := stageBundle(data, b, dest); err != nil {
		t.Fatalf("stageBundle: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "libpdfium.so")); string(got) != "PDFIUMBYTES" {
		t.Errorf("pdfium contents = %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "eng.traineddata")); string(got) != "OCRDATA" {
		t.Errorf("tessdata contents = %q", got)
	}
	// lit binary is executable...
	litPath := filepath.Join(dest, "lit")
	if info, err := os.Stat(litPath); err != nil || info.Mode().Perm()&0o100 == 0 {
		t.Errorf("lit not executable: mode=%v err=%v", info.Mode(), err)
	}
	// ...and decoys/nested paths are NOT written.
	if fileExists(filepath.Join(dest, "LICENSE")) || fileExists(filepath.Join(dest, "BUNDLE.txt")) {
		t.Error("decoy files should not be extracted")
	}
	if _, err := os.Stat(filepath.Join(dest, "sub")); !os.IsNotExist(err) {
		t.Error("nested decoy path should not be extracted")
	}
}

func TestStageBundleZip(t *testing.T) {
	data := makeZipBundle(t, map[string]string{
		"lit.exe":         "MZlit",
		"pdfium.dll":      "PDFIUMDLL",
		"eng.traineddata": "OCRDATA",
	})
	b := litBundle{Asset: "lit-windows-amd64.zip", LitBin: "lit.exe", Pdfium: "pdfium.dll", Tessdata: "eng.traineddata"}

	dest := filepath.Join(t.TempDir(), "lit-v9.9.9", "windows-amd64")
	if err := stageBundle(data, b, dest); err != nil {
		t.Fatalf("stageBundle: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "lit.exe")); string(got) != "MZlit" {
		t.Errorf("lit.exe contents = %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "pdfium.dll")); string(got) != "PDFIUMDLL" {
		t.Errorf("pdfium.dll contents = %q", got)
	}
}

func TestStageBundleMissingMember(t *testing.T) {
	// Archive has lit + pdfium but not the expected tessdata.
	data := makeTarGzBundle(t, map[string]string{"lit": "x", "libpdfium.so": "y"})
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "eng.traineddata"}

	dest := filepath.Join(t.TempDir(), "lit", "linux-amd64")
	err := stageBundle(data, b, dest)
	if err == nil || !contains(err.Error(), "missing expected member") {
		t.Fatalf("expected missing-member error, got %v", err)
	}
	if fileExists(filepath.Join(dest, "lit")) {
		t.Error("nothing should be installed when a member is missing")
	}
}

func TestEnsureBundleFromDownloadVerifiesChecksum(t *testing.T) {
	data := ocrTarGz(t)
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", SHA256: sha256hex(data), LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "eng.traineddata"}

	var gotURL string
	calls := 0
	fetch := func(url string) ([]byte, error) { gotURL = url; calls++; return data, nil }

	// Force a known cache dir via the env override path is not enough (it
	// targets UserCacheDir); call the function directly with a temp HOME.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	dir, err := ensureBundleFromDownload("lit-v9.9.9", b, "https://example.test/dl", fetch)
	if err != nil {
		t.Fatalf("ensureBundleFromDownload: %v", err)
	}
	if want := "https://example.test/dl/lit-v9.9.9/lit-linux-amd64.tar.gz"; gotURL != want {
		t.Errorf("url = %q, want %q", gotURL, want)
	}
	if !fileExists(filepath.Join(dir, "lit")) || !fileExists(filepath.Join(dir, "eng.traineddata")) {
		t.Error("expected lit + tessdata staged")
	}
	// Second call reuses the cache (no second fetch).
	if _, err := ensureBundleFromDownload("lit-v9.9.9", b, "https://example.test/dl", fetch); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("fetch called %d times, want 1 (second call should reuse cache)", calls)
	}
}

func TestEnsureBundleFromDownloadChecksumMismatch(t *testing.T) {
	data := ocrTarGz(t)
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", SHA256: "deadbeef", LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "eng.traineddata"}
	fetch := func(string) ([]byte, error) { return data, nil }
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	_, err := ensureBundleFromDownload("lit-v9.9.9", b, "https://example.test/dl", fetch)
	if err == nil || !contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}
}

func TestEnsureBundleFromBytesEmbedded(t *testing.T) {
	data := ocrTarGz(t)
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "eng.traineddata"}
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	dir, err := ensureBundleFromBytes("embedded-test", data, b)
	if err != nil {
		t.Fatalf("ensureBundleFromBytes: %v", err)
	}
	for _, m := range []string{"lit", "libpdfium.so", "eng.traineddata"} {
		if !fileExists(filepath.Join(dir, m)) {
			t.Errorf("embedded member %q not staged", m)
		}
	}
	// Idempotent reuse.
	dir2, err := ensureBundleFromBytes("embedded-test", data, b)
	if err != nil || dir2 != dir {
		t.Fatalf("second call dir=%q err=%v (want stable, no error)", dir2, err)
	}
}

func TestLitCLIFromDirWiresTessdataAndPdfium(t *testing.T) {
	dir := t.TempDir()
	withOCR := litBundle{LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "eng.traineddata"}
	c := litCLIFromDir(withOCR, dir)
	if c.pdfiumDir != dir || c.tessdataDir != dir {
		t.Fatalf("pdfiumDir=%q tessdataDir=%q, want both %q", c.pdfiumDir, c.tessdataDir, dir)
	}
	// OCR parse args should carry --tessdata-path pointing at the bundle dir.
	args := c.parseArgs("doc.pdf", defaultParseOptions())
	if !hasFlagValue(args, "--tessdata-path", dir) {
		t.Errorf("parseArgs missing --tessdata-path %q: %v", dir, args)
	}

	// A no-OCR bundle (windows/arm64) must not set tessdata.
	noOCR := litBundle{LitBin: "lit.exe", Pdfium: "pdfium.dll", Tessdata: ""}
	c2 := litCLIFromDir(noOCR, dir)
	if c2.tessdataDir != "" {
		t.Errorf("tessdataDir = %q, want empty for no-OCR bundle", c2.tessdataDir)
	}
}

func TestParseArgsNoTessdataWhenOCRDisabled(t *testing.T) {
	c := litCLIFromDir(litBundle{LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "eng.traineddata"}, t.TempDir())
	opt := defaultParseOptions()
	opt.OCR = false
	args := c.parseArgs("doc.pdf", opt)
	for _, a := range args {
		if a == "--tessdata-path" {
			t.Errorf("--tessdata-path should not appear when OCR disabled: %v", args)
		}
	}
}

func hasFlagValue(args []string, flag, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool { return bytes.Contains([]byte(s), []byte(sub)) }
