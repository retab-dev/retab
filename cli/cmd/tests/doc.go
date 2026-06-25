// Package tests holds cross-command conformance tests for the Retab CLI.
//
// These tests assert invariants that every command in a family must satisfy —
// e.g. every primitive list forwards its file/date filters, every list with
// --before/--after enforces their mutual exclusion, every list renderer honors
// --output csv, and every delete gates on --yes. They discover the relevant
// commands by walking cmd.RootCommand(), so a newly added command that breaks a
// rule fails the suite without anyone remembering to add a bespoke test.
//
// The package intentionally lives outside package cmd: it exercises the CLI
// through the same exported surface external tooling would use.
package tests
