package cmd

import (
	"strings"
	"testing"
)

// TestResolveDownloadDest pins the destination-resolution rules for
// `retab files download`: positional vs. -o flag, "-" → stdout, and the
// mutex error when both forms are passed.
//
// The CLI surfaces two equivalent destination inputs because users coming
// from the help text (which used to say "- for stdout") expect the
// positional form to work, while older docs and pipelines lean on the
// -o flag. This test guards both shapes from regressing.
func TestResolveDownloadDest(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		oFlag       string
		wantPath    string
		wantStdout  bool
		wantErr     bool
		wantErrSubs string
	}{
		{
			name:       "one arg, no flag — defer to server filename",
			args:       []string{"file_abc"},
			oFlag:      "",
			wantPath:   "",
			wantStdout: false,
		},
		{
			name:       "one arg, explicit -o path",
			args:       []string{"file_abc"},
			oFlag:      "./out.pdf",
			wantPath:   "./out.pdf",
			wantStdout: false,
		},
		{
			name:       "one arg, -o - means stdout",
			args:       []string{"file_abc"},
			oFlag:      "-",
			wantPath:   "",
			wantStdout: true,
		},
		{
			name:       "two args, positional - means stdout",
			args:       []string{"file_abc", "-"},
			oFlag:      "",
			wantPath:   "",
			wantStdout: true,
		},
		{
			name:       "two args, positional path",
			args:       []string{"file_abc", "./out.pdf"},
			oFlag:      "",
			wantPath:   "./out.pdf",
			wantStdout: false,
		},
		{
			name:        "both positional and flag — reject",
			args:        []string{"file_abc", "./out.pdf"},
			oFlag:       "./other",
			wantErr:     true,
			wantErrSubs: "cannot use positional ./out.pdf and -o flag together",
		},
		{
			name:        "both positional - and flag — also reject",
			args:        []string{"file_abc", "-"},
			oFlag:       "./other",
			wantErr:     true,
			wantErrSubs: "cannot use positional - and -o flag together",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotStdout, err := resolveDownloadDest(tc.args, tc.oFlag)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got path=%q stdout=%v", gotPath, gotStdout)
				}
				if tc.wantErrSubs != "" && !strings.Contains(err.Error(), tc.wantErrSubs) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrSubs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotPath != tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tc.wantPath)
			}
			if gotStdout != tc.wantStdout {
				t.Errorf("toStdout = %v, want %v", gotStdout, tc.wantStdout)
			}
		})
	}
}

// TestFilesDownloadCmdArgsRange pins that the cobra Args validator accepts
// 1-2 positional args and rejects 0 or 3+. This catches accidental
// regressions to cobra.ExactArgs(1) or RangeArgs(1, 3).
func TestFilesDownloadCmdArgsRange(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "zero args", args: []string{}, wantErr: true},
		{name: "one arg", args: []string{"file_abc"}, wantErr: false},
		{name: "two args", args: []string{"file_abc", "-"}, wantErr: false},
		{name: "two args path", args: []string{"file_abc", "./out.pdf"}, wantErr: false},
		{name: "three args", args: []string{"file_abc", "-", "extra"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := filesDownloadCmd.Args(filesDownloadCmd, tc.args)
			if tc.wantErr && err == nil {
				t.Fatalf("want error for args=%v, got nil", tc.args)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for args=%v: %v", tc.args, err)
			}
		})
	}
}
