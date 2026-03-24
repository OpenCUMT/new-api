package common

import "testing"

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  USER@OpenCUMT.org "); got != "user@opencumt.org" {
		t.Fatalf("expected normalized email, got %q", got)
	}
}

func TestValidateRestrictedRegisterEmail(t *testing.T) {
	testCases := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{name: "valid email", email: "member@opencumt.org", wantErr: false},
		{name: "valid mixed case", email: "Member@OpenCUMT.org", wantErr: false},
		{name: "invalid domain", email: "member@example.com", wantErr: true},
		{name: "empty email", email: "", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRestrictedRegisterEmail(tc.email)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q", tc.email)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.email, err)
			}
		})
	}
}
