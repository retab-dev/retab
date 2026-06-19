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

// Self-contained `lit` bundles
// ----------------------------
// The local-first `files parse|grep|inspect` commands shell out to LiteParse's
// `lit` CLI. `lit` is a Rust+C++ binary that dlopen()s libpdfium at runtime and
// reads Tesseract OCR data (Latin.traineddata) from disk — none of which fits in
// our pure-Go, CGO-free `retab` binary. So we ship `lit` as a bundle of three
// files (lit + libpdfium + Latin.traineddata) and resolve it one of two ways:
//
//   1. Embedded (default release build, -tags retab_embed_lit): the bundle for
//      the current platform is baked into `retab` via go:embed and unpacked to
//      the user cache dir on first use. Fully offline, no network, no API.
//   2. Downloaded (source builds without the embed tag): the bundle is fetched
//      from the pinned GitHub release, checksum-verified, and cached.
//
// Either way it lands in the same cache layout and `lit` is exec'd with
// PDFIUM_LIB_PATH (so it finds libpdfium) and --tessdata-path (so OCR finds its
// language data). See litCLI.command / parseArgs.

// litBundleTag pins the bundle release this CLI build expects. It matches the
// tag published by build-liteparse.yml and the assets goreleaser embeds. Keep
// in sync with LIT_BUNDLE_TAG in scripts/fetch-lit-assets.sh.
//
// -r2: re-bundle of the crates-v2.0.3 `lit` source carrying the multilingual
// Latin-script tessdata model (Latin.traineddata) instead of the original
// English-only eng.traineddata. The lit binary is unchanged.
const litBundleTag = "lit-v2.0.3-r2"

// litBundleBaseURL is the download root for the fallback (non-embedded) path.
// Overridable via RETAB_LITEPARSE_BUNDLE_URL for mirrors and tests.
const litBundleBaseURL = "https://github.com/retab-dev/retab/releases/download"

// errBundleUnavailable signals that neither an embedded bundle nor a
// checksum-pinned downloadable bundle exists for the current platform, so the
// caller should fall back to the manual install hint.
var errBundleUnavailable = errors.New("no managed liteparse bundle for this platform")

// litBundle describes one platform's lit bundle: the members packed in the
// archive plus (for the download path) the archive's expected checksum.
type litBundle struct {
	Asset    string // archive filename, e.g. lit-darwin-arm64.tar.gz
	SHA256   string // hex sha256 of the archive; empty => download disabled
	LitBin   string // binary name inside the archive (lit or lit.exe)
	Pdfium   string // pdfium shared-library filename inside the archive
	Tessdata string // OCR data filename (Latin.traineddata); empty if no OCR
}

// litBundles is the per-platform manifest, keyed by GOOS/GOARCH.
//
// SHA256 only gates the *download* fallback (we refuse to run an unverified
// download); the embedded path is compiled in and trusted, so it works
// regardless of SHA256. The values are filled from the checksums.txt that
// build-liteparse.yml publishes. windows/arm64 ships OCR-less (no Tessdata),
// matching the build matrix.
var litBundles = map[string]litBundle{
	"linux/amd64":   {Asset: "lit-linux-amd64.tar.gz", SHA256: "3f89316605b3bc550fb58c3fd3d1de5362e486854e4eaa6955aa73acb603465b", LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "Latin.traineddata"},
	"linux/arm64":   {Asset: "lit-linux-arm64.tar.gz", SHA256: "f55738886dbbfe9c0829c296c040d36ce094186e294f04915dc7ecbe9313d244", LitBin: "lit", Pdfium: "libpdfium.so", Tessdata: "Latin.traineddata"},
	"darwin/amd64":  {Asset: "lit-darwin-amd64.tar.gz", SHA256: "155636a80a246520b745c87adfa0a8a7ed93847c77479c4bfa8fb77fbe3652e5", LitBin: "lit", Pdfium: "libpdfium.dylib", Tessdata: "Latin.traineddata"},
	"darwin/arm64":  {Asset: "lit-darwin-arm64.tar.gz", SHA256: "b4eaab56fa09fa08fc3bcbfd25013c79e38dfb77947f9c8983065350df6bd3f5", LitBin: "lit", Pdfium: "libpdfium.dylib", Tessdata: "Latin.traineddata"},
	"windows/amd64": {Asset: "lit-windows-amd64.tar.gz", SHA256: "38ba9f10a33ff9a2172b5138182e4d1baba5ddd7232b21b633ad79bcb99b42eb", LitBin: "lit.exe", Pdfium: "pdfium.dll", Tessdata: "Latin.traineddata"},
	"windows/arm64": {Asset: "lit-windows-arm64.tar.gz", SHA256: "546c3035ed6c30b3e4569432ea14eff07fd81d0748973ba4d6b8860752c32511", LitBin: "lit.exe", Pdfium: "pdfium.dll", Tessdata: ""},
}

func currentBundle() (litBundle, bool) {
	b, ok := litBundles[runtime.GOOS+"/"+runtime.GOARCH]
	return b, ok
}

// wantMembers is the set of archive members to extract for a bundle. Other
// files (LICENSE, BUNDLE.txt, nested paths) are ignored.
func (b litBundle) wantMembers() map[string]bool {
	want := map[string]bool{b.LitBin: true, b.Pdfium: true}
	if b.Tessdata != "" {
		want[b.Tessdata] = true
	}
	return want
}

// litBundleFetch downloads the bytes at url. It's a package var so tests can
// inject a local fixture instead of hitting the network.
var litBundleFetch = httpFetch

func httpFetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
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

func liteParseBinCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "retab", "liteparse-bin"), nil
}

// resolveManagedLit returns a litCLI backed by the embedded bundle if this
// build has one, otherwise by a checksum-verified download. errBundleUnavailable
// means neither is available for this platform.
func resolveManagedLit() (*litCLI, error) {
	b, ok := currentBundle()
	if !ok {
		return nil, errBundleUnavailable
	}

	// Embedded path (default release build): trusted, compiled-in, offline.
	if data, ok := embeddedLitArchive(); ok && len(data) > 0 {
		dir, err := ensureBundleFromBytes("embedded-"+litBundleTag, data, b)
		if err != nil {
			return nil, err
		}
		return litCLIFromDir(b, dir), nil
	}

	// Download fallback (source builds): requires a pinned checksum.
	if b.SHA256 == "" {
		return nil, errBundleUnavailable
	}
	dir, err := ensureBundleFromDownload(litBundleTag, b, bundleBaseURL(), litBundleFetch)
	if err != nil {
		return nil, err
	}
	return litCLIFromDir(b, dir), nil
}

// litCLIFromDir builds a litCLI pointed at an extracted bundle dir, wiring the
// pdfium and (when present) tessdata directories.
func litCLIFromDir(b litBundle, dir string) *litCLI {
	c := &litCLI{bin: filepath.Join(dir, b.LitBin), pdfiumDir: dir}
	if b.Tessdata != "" {
		c.tessdataDir = dir
	}
	return c
}

type fetchFunc func(url string) ([]byte, error)

// ensureBundleFromDownload fetches <baseURL>/<tag>/<asset>, verifies its sha256
// against b.SHA256, and stages it into the cache. Returns the extracted dir.
func ensureBundleFromDownload(tag string, b litBundle, baseURL string, fetch fetchFunc) (string, error) {
	base, err := liteParseBinCacheDir()
	if err != nil {
		return "", err
	}
	destDir := filepath.Join(base, tag, runtime.GOOS+"-"+runtime.GOARCH)
	if bundleStaged(destDir, b) {
		return destDir, nil
	}

	url := strings.TrimRight(baseURL, "/") + "/" + tag + "/" + b.Asset
	data, err := fetch(url)
	if err != nil {
		return "", fmt.Errorf("download liteparse bundle: %w", err)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); !strings.EqualFold(got, b.SHA256) {
		return "", fmt.Errorf("liteparse bundle checksum mismatch for %s: got %s, want %s", b.Asset, got, b.SHA256)
	}
	if err := stageBundle(data, b, destDir); err != nil {
		return "", err
	}
	return destDir, nil
}

// ensureBundleFromBytes stages an in-memory (embedded) bundle into the cache.
// The cache key folds in len(data) so a CLI upgrade carrying a different bundle
// re-extracts instead of reusing stale files; no checksum is needed because the
// bytes are compiled into this binary.
func ensureBundleFromBytes(key string, data []byte, b litBundle) (string, error) {
	base, err := liteParseBinCacheDir()
	if err != nil {
		return "", err
	}
	destDir := filepath.Join(base, fmt.Sprintf("%s-%d", key, len(data)), runtime.GOOS+"-"+runtime.GOARCH)
	if bundleStaged(destDir, b) {
		return destDir, nil
	}
	if err := stageBundle(data, b, destDir); err != nil {
		return "", err
	}
	return destDir, nil
}

// bundleStaged reports whether destDir already holds every member of the bundle.
func bundleStaged(destDir string, b litBundle) bool {
	for name := range b.wantMembers() {
		if !fileExists(filepath.Join(destDir, name)) {
			return false
		}
	}
	return true
}

// stageBundle extracts the wanted members of a .tar.gz (or .zip) archive into
// destDir atomically: it extracts into a temp dir, verifies all members landed,
// makes the binary executable, then renames into place so a crash mid-extract
// never leaves a half-populated cache dir.
func stageBundle(data []byte, b litBundle, destDir string) error {
	if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(destDir), ".lit-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	want := b.wantMembers()
	if strings.HasSuffix(b.Asset, ".zip") {
		err = extractZipMembers(data, tmpDir, want)
	} else {
		err = extractTarGzMembers(data, tmpDir, want)
	}
	if err != nil {
		return err
	}
	for name := range want {
		if !fileExists(filepath.Join(tmpDir, name)) {
			return fmt.Errorf("liteparse bundle %s missing expected member %q", b.Asset, name)
		}
	}
	if err := os.Chmod(filepath.Join(tmpDir, b.LitBin), 0o755); err != nil {
		return err
	}

	if err := os.Rename(tmpDir, destDir); err != nil {
		// A concurrent invocation may have populated destDir first.
		if bundleStaged(destDir, b) {
			return nil
		}
		return fmt.Errorf("install liteparse bundle: %w", err)
	}
	return nil
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
