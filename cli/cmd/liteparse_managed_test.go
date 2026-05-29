package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

func TestInstallLitBundleTarGz(t *testing.T) {
	data := makeTarGzBundle(t, map[string]string{
		"lit":          "#!/bin/sh\necho lit",
		"libpdfium.so": "PDFIUMBYTES",
	})
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", SHA256: sha256hex(data), LitBin: "lit", Pdfium: "libpdfium.so"}

	var gotURL string
	fetch := func(url string) ([]byte, error) { gotURL = url; return data, nil }

	dest := filepath.Join(t.TempDir(), "lit-v9.9.9", "linux-amd64")
	litPath, pdfiumDir, err := installLitBundle(b, "lit-v9.9.9", "https://example.test/dl", dest, fetch)
	if err != nil {
		t.Fatalf("installLitBundle: %v", err)
	}
	if want := "https://example.test/dl/lit-v9.9.9/lit-linux-amd64.tar.gz"; gotURL != want {
		t.Errorf("url = %q, want %q", gotURL, want)
	}
	if litPath != filepath.Join(dest, "lit") {
		t.Errorf("litPath = %q", litPath)
	}
	if pdfiumDir != dest {
		t.Errorf("pdfiumDir = %q, want %q", pdfiumDir, dest)
	}
	// Wanted members extracted...
	if got, _ := os.ReadFile(filepath.Join(dest, "libpdfium.so")); string(got) != "PDFIUMBYTES" {
		t.Errorf("pdfium contents = %q", got)
	}
	// ...lit binary is executable...
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

func TestInstallLitBundleZip(t *testing.T) {
	data := makeZipBundle(t, map[string]string{
		"lit.exe":    "MZlit",
		"pdfium.dll": "PDFIUMDLL",
	})
	b := litBundle{Asset: "lit-windows-amd64.zip", SHA256: sha256hex(data), LitBin: "lit.exe", Pdfium: "pdfium.dll"}
	fetch := func(string) ([]byte, error) { return data, nil }

	dest := filepath.Join(t.TempDir(), "lit-v9.9.9", "windows-amd64")
	litPath, _, err := installLitBundle(b, "lit-v9.9.9", "https://example.test/dl", dest, fetch)
	if err != nil {
		t.Fatalf("installLitBundle: %v", err)
	}
	if got, _ := os.ReadFile(litPath); string(got) != "MZlit" {
		t.Errorf("lit.exe contents = %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "pdfium.dll")); string(got) != "PDFIUMDLL" {
		t.Errorf("pdfium.dll contents = %q", got)
	}
}

func TestInstallLitBundleChecksumMismatch(t *testing.T) {
	data := makeTarGzBundle(t, map[string]string{"lit": "x", "libpdfium.so": "y"})
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", SHA256: "deadbeef", LitBin: "lit", Pdfium: "libpdfium.so"}
	fetch := func(string) ([]byte, error) { return data, nil }

	dest := filepath.Join(t.TempDir(), "lit", "linux-amd64")
	_, _, err := installLitBundle(b, "lit-v9.9.9", "https://example.test/dl", dest, fetch)
	if err == nil || !contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}
	if fileExists(filepath.Join(dest, "lit")) {
		t.Error("nothing should be installed on checksum failure")
	}
}

func TestInstallLitBundleMissingMember(t *testing.T) {
	// Archive has lit but not the expected pdfium lib.
	data := makeTarGzBundle(t, map[string]string{"lit": "x"})
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", SHA256: sha256hex(data), LitBin: "lit", Pdfium: "libpdfium.so"}
	fetch := func(string) ([]byte, error) { return data, nil }

	dest := filepath.Join(t.TempDir(), "lit", "linux-amd64")
	_, _, err := installLitBundle(b, "lit-v9.9.9", "https://example.test/dl", dest, fetch)
	if err == nil || !contains(err.Error(), "missing expected member") {
		t.Fatalf("expected missing-member error, got %v", err)
	}
}

func TestInstallLitBundleReusesCache(t *testing.T) {
	data := makeTarGzBundle(t, map[string]string{"lit": "v1", "libpdfium.so": "p1"})
	b := litBundle{Asset: "lit-linux-amd64.tar.gz", SHA256: sha256hex(data), LitBin: "lit", Pdfium: "libpdfium.so"}

	calls := 0
	fetch := func(string) ([]byte, error) { calls++; return data, nil }

	dest := filepath.Join(t.TempDir(), "lit-v9.9.9", "linux-amd64")
	if _, _, err := installLitBundle(b, "lit-v9.9.9", "https://example.test/dl", dest, fetch); err != nil {
		t.Fatal(err)
	}
	// Second call must hit the cache, not fetch again.
	if _, _, err := installLitBundle(b, "lit-v9.9.9", "https://example.test/dl", dest, fetch); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("fetch called %d times, want 1 (second call should reuse cache)", calls)
	}
}

func TestEnsureManagedLitUnavailableWhenNoChecksum(t *testing.T) {
	// All shipped manifest entries start with empty SHA256, so until checksums
	// are pinned post-CI, every platform must report errBundleUnavailable.
	_, _, err := ensureManagedLit(litBundleTag)
	if !errors.Is(err, errBundleUnavailable) {
		t.Fatalf("expected errBundleUnavailable while checksums are unpinned, got %v", err)
	}
}

func contains(s, sub string) bool { return bytes.Contains([]byte(s), []byte(sub)) }
