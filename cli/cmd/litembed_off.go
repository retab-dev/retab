//go:build !retab_embed_lit

package cmd

// embeddedLitArchive returns the lit bundle baked into this binary, or false
// when none is embedded.
//
// This stub is compiled for builds *without* -tags retab_embed_lit (e.g.
// `go install ./cli` from source, or local `go build`/`go test`). In that case
// the resolver falls back to downloading the checksum-pinned bundle. The
// release build (.goreleaser.yaml) sets -tags retab_embed_lit, which swaps in
// the per-platform litembed_<os>_<arch>.go file that actually embeds the asset.
func embeddedLitArchive() ([]byte, bool) { return nil, false }
