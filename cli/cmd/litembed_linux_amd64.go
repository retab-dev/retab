//go:build retab_embed_lit

package cmd

import _ "embed"

// The self-contained lit bundle for this platform (lit + libpdfium + OCR data),
// baked in at release time. The asset is staged into assets/ by
// scripts/fetch-lit-assets.sh before the goreleaser build (it is gitignored, so
// it is absent in normal checkouts — only the retab_embed_lit release build,
// which constrains this file, references it).
//
//go:embed assets/lit-linux-amd64.tar.gz
var embeddedLitData []byte

func embeddedLitArchive() ([]byte, bool) { return embeddedLitData, true }
