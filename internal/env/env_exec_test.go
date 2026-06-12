//go:build !wasip1

package env

import "testing"

func TestParseGoMinor(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"standard", "go version go1.24.2 darwin/arm64", "1.24"},
		{"linux amd64", "go version go1.21.0 linux/amd64", "1.21"},
		{"no patch", "go version go1.24 windows/amd64", "1.24"},
		{"major only", "go version go1 plan9/386", "1"},
		{"empty", "", ""},
		{"too short", "go version", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseGoMinor(tt.raw); got != tt.want {
				t.Errorf("parseGoMinor(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestExecEnvGetwd(t *testing.T) {
	wd, err := New().Getwd()
	if err != nil {
		t.Fatalf("Getwd() returned error: %v", err)
	}
	if wd == "" {
		t.Fatal("Getwd() returned an empty path")
	}
}

func TestExecEnvImplementsEnvironment(t *testing.T) {
	var _ Environment = New()
}
