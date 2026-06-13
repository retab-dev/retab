package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// writeImageFile creates an empty file so collectRenderedPages can see it.
func writeImageFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestPageNumberFromImagePath(t *testing.T) {
	cases := map[string]struct {
		want int
		ok   bool
	}{
		"page_1.png":          {1, true},
		"page_2.png":          {2, true},
		"page_10.jpg":         {10, true},
		"page-3.png":          {3, true},
		"/tmp/out/page_7.png": {7, true},
		"thumbnail.png":       {0, false},
		"page_.png":           {0, false},
	}
	for name, want := range cases {
		got, ok := pageNumberFromImagePath(name)
		if ok != want.ok || got != want.want {
			t.Errorf("pageNumberFromImagePath(%q) = (%d, %v), want (%d, %v)", name, got, ok, want.want, want.ok)
		}
	}
}

// TestCollectRenderedPagesLabelsRealPageNumbers guards the bug where rendering
// pages 2-3 reported page:1 -> page_2.png and page:2 -> page_3.png because the
// page number was the file's ordinal position, not its real page number.
func TestCollectRenderedPagesLabelsRealPageNumbers(t *testing.T) {
	dir := t.TempDir()
	writeImageFile(t, dir, "page_2.png")
	writeImageFile(t, dir, "page_3.png")

	pages, err := collectRenderedPages(dir, "2-3")
	if err != nil {
		t.Fatalf("collectRenderedPages: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d: %+v", len(pages), pages)
	}
	if pages[0].Page != 2 || filepath.Base(pages[0].Path) != "page_2.png" {
		t.Errorf("first page mislabeled: got page=%d path=%s, want page=2 page_2.png", pages[0].Page, filepath.Base(pages[0].Path))
	}
	if pages[1].Page != 3 || filepath.Base(pages[1].Path) != "page_3.png" {
		t.Errorf("second page mislabeled: got page=%d path=%s, want page=3 page_3.png", pages[1].Page, filepath.Base(pages[1].Path))
	}
}

// TestCollectRenderedPagesReportsAlreadyPresentRequestedPage guards the bug
// where re-rendering a page into a directory that already held its PNG returned
// an empty (null) page list because only newly written files were reported.
func TestCollectRenderedPagesReportsAlreadyPresentRequestedPage(t *testing.T) {
	dir := t.TempDir()
	writeImageFile(t, dir, "page_1.png") // pre-existing from a prior render

	pages, err := collectRenderedPages(dir, "1")
	if err != nil {
		t.Fatalf("collectRenderedPages: %v", err)
	}
	if len(pages) != 1 || pages[0].Page != 1 {
		t.Fatalf("expected page 1 reported even though it already existed, got %+v", pages)
	}
}

// TestCollectRenderedPagesExcludesUnrequestedPages ensures images for pages not
// in the target spec (e.g. left in the dir by an earlier render) are not
// reported as part of this render.
func TestCollectRenderedPagesExcludesUnrequestedPages(t *testing.T) {
	dir := t.TempDir()
	writeImageFile(t, dir, "page_1.png") // stale, from an earlier render
	writeImageFile(t, dir, "page_5.png")
	writeImageFile(t, dir, "page_6.png")

	pages, err := collectRenderedPages(dir, "5,6")
	if err != nil {
		t.Fatalf("collectRenderedPages: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected only the 2 requested pages, got %d: %+v", len(pages), pages)
	}
	if pages[0].Page != 5 || pages[1].Page != 6 {
		t.Errorf("wrong pages reported: %+v", pages)
	}
}
