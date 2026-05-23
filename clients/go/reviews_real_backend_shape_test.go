package retab

import "testing"

func TestReviewWireConstantsMatchHardCutover(t *testing.T) {
	cases := []struct {
		wire     string
		constant string
	}{
		{"accepted", string(SubmissionStatusAccepted)},
		{"pending", string(ResumeStatusPending)},
		{"resumed", string(ResumeStatusResumed)},
		{"skipped", string(ResumeStatusSkipped)},
	}
	for _, tc := range cases {
		if tc.constant != tc.wire {
			t.Errorf("constant %q does not match wire value %q", tc.constant, tc.wire)
		}
	}
}
