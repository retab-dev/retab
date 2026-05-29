package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Managed `lit` bundles
// ----------------------
// Upstream LiteParse ships release tarballs for only 3 of the 6 platforms
// Retab supports, and those tarballs contain only the `lit` binary — not the
// libpdfium that `lit` dlopen()s at runtime. So when no `lit` is on PATH, the
// Retab CLI downloads a self-contained bundle we build + publish ourselves
// (see .github/workflows/build-liteparse.yml + scripts/build-liteparse.sh):
// `lit` packaged together with the matching libpdfium. The bundle is fetched
// once, checksum-verified against the pinned manifest below, extracted into the
// user cache dir, and reused thereafter. PDFIUM_LIB_PATH is set to the extracted
// directory when exec'ing the managed `lit` (see litCLI.command).

// litBundleTag pins the bundle release this CLI build expects. It matches the
// tag published by build-liteparse.yml.
const litBundleTag = "lit-v2.0.3"

// litBundleBaseURL is the release download root. Overridable via
// RETAB_LITEPARSE_BUNDLE_URL for mirrors and tests.
const litBundleBaseURL = "https://github.com/retab-dev/retab/releases/download"

// errBundleUnavailable signals that no checksum-pinned bundle exists for the
// current platform, so the caller should fall back to the install hint rather
// than treat it as a hard download failure.
var errBundleUnavailable = errors.New("no managed liteparse bundle for this platform")

// litBundle describes one platform's downloadable lit+pdfium archive.
type litBundle struct {
	Asset  string // archive filename, e.g. lit-darwin-arm64.tar.gz
	SHA256 string // hex sha256 of the archive; empty => not yet pinned (disabled)
	LitBin string // binary name inside the archive (lit or lit.exe)
	Pdfium string // pdfium shared-library filename inside the archive
}

// litBundles is the pinned per-platform manifest, keyed by GOOS/GOARCH.
//
// SHA256 values come from the checksums.txt that build-liteparse.yml publishes.
// They are intentionally empty until that workflow has run for litBundleTag:
// an empty SHA256 disables managed download for the platform (we refuse to run
// an unverified binary) and the resolver falls back to the manual install hint.
// Filling these six values is the only step needed to activate download-on-demand.
var litBundles = map[string]litBundle{
	"linux/amd64":   {Asset: "lit-linux-amd64.tar.gz", SHA256: "", LitBin: "lit", Pdfium: "libpdfium.so"},
	"linux/arm64":   {Asset: "lit-linux-arm64.tar.gz", SHA256: "", LitBin: "lit", Pdfium: "libpdfium.so"},
	"darwin/amd64":  {Asset: "lit-darwin-amd64.tar.gz", SHA256: "", LitBin: "lit", Pdfium: "libpdfium.dylib"},
	"darwin/arm64":  {Asset: "lit-darwin-arm64.tar.gz", SHA256: "", LitBin: "lit", Pdfium: "libpdfium.dylib"},
	"windows/amd64": {Asset: "lit-windows-amd64.zip", SHA256: "", LitBin: "lit.exe", Pdfium: "pdfium.dll"},
	"windows/arm64": {Asset: "lit-windows-arm64.zip", SHA256: "", LitBin: "lit.exe", Pdfium: "pdfium.dll"},
}

// litBundleFetch downloads the bytes at url. It's a package var so tests can
// inject a local fixture instead of hitting the network.
var litBundleFetch = httpFetch

func httpFetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func bundleBaseURL() string {
	if v := os.Getenv("RETAB_LITEPARSE_BUNDLE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return litBundleBaseURL
}

// ensureManagedLit resolves (downloading + verifying + extracting on first use)
// the managed `lit` bundle for the current platform at the given tag. It
// returns the path to the `lit` binary and the directory holding libpdfium
// (to be exported as PDFIUM_LIB_PATH). errBundleUnavailable is returned when no
// checksum-pinned bundle exists for this platform.
func ensureManagedLit(tag string) (litPath, pdfiumDir string, err error) {
	platform := runtime.GOOS + "/" + runtime.GOARCH
	b, ok := litBundles[platform]
	if !ok || b.SHA256 == "" {
		return "", "", errBundleUnavailable
	}
	base, err := liteParseBinCacheDir()
	if err != nil {
		return "", "", err
	}
	destDir := filepath.Join(base, tag, runtime.GOOS+"-"+runtime.GOARCH)
	return installLitBundle(b, tag, bundleBaseURL(), destDir, litBundleFetch)
}

func liteParseBinCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "retab", "liteparse-bin"), nil
}

type fetchFunc func(url string) ([]byte, error)

// installLitBundle is the testable core: if destDir already holds the binary,
// it's reused; otherwise the archive is fetched from
// <baseURL>/<tag>/<asset>, its sha256 verified against b.SHA256, and the lit
// binary + pdfium library extracted into destDir.
func installLitBundle(b litBundle, tag, baseURL, destDir string, fetch fetchFunc) (litPath, pdfiumDir string, err error) {
	litPath = filepath.Join(destDir, b.LitBin)
	pdfiumPath := filepath.Join(destDir, b.Pdfium)
	if fileExists(litPath) && fileExists(pdfiumPath) {
		return litPath, destDir, nil
	}

	url := strings.TrimRight(baseURL, "/") + "/" + tag + "/" + b.Asset
	data, err := fetch(url)
	if err != nil {
		return "", "", fmt.Errorf("download liteparse bundle: %w", err)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); !strings.EqualFold(got, b.SHA256) {
		return "", "", fmt.Errorf("liteparse bundle checksum mismatch for %s: got %s, want %s", b.Asset, got, b.SHA256)
	}

	// Extract into a temp dir first, then atomically rename into place so a
	// crash mid-extract never leaves a half-populated cache dir.
	if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
		return "", "", err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(destDir), ".lit-extract-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmpDir)

	want := map[string]bool{b.LitBin: true, b.Pdfium: true}
	if strings.HasSuffix(b.Asset, ".zip") {
		err = extractZipMembers(data, tmpDir, want)
	} else {
		err = extractTarGzMembers(data, tmpDir, want)
	}
	if err != nil {
		return "", "", err
	}
	for name := range want {
		if !fileExists(filepath.Join(tmpDir, name)) {
			return "", "", fmt.Errorf("liteparse bundle %s missing expected member %q", b.Asset, name)
		}
	}
	if err := os.Chmod(filepath.Join(tmpDir, b.LitBin), 0o755); err != nil {
		return "", "", err
	}

	// If a concurrent invocation already populated destDir, keep theirs.
	if err := os.Rename(tmpDir, destDir); err != nil {
		if fileExists(litPath) && fileExists(pdfiumPath) {
			return litPath, destDir, nil
		}
		return "", "", fmt.Errorf("install liteparse bundle: %w", err)
	}
	return litPath, destDir, nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// extractTarGzMembers writes the named top-level members (by base name) from a
// gzip'd tar archive into destDir. Unwanted members and any path traversal are
// ignored.
func extractTarGzMembers(data []byte, destDir string, want map[string]bool) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("open bundle gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read bundle tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		name := filepath.Base(hdr.Name)
		if !want[name] {
			continue
		}
		if err := writeMember(filepath.Join(destDir, name), tr); err != nil {
			return err
		}
	}
	return nil
}

// extractZipMembers is the zip counterpart of extractTarGzMembers.
func extractZipMembers(data []byte, destDir string, want map[string]bool) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open bundle zip: %w", err)
	}
	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if f.FileInfo().IsDir() || !want[name] {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip member %q: %w", name, err)
		}
		err = writeMember(filepath.Join(destDir, name), rc)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func writeMember(path string, r io.Reader) error {
	out, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
