package apply

import "testing"

func TestNormalizeTimeout(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "5s", want: "5000ms"},
		{input: "10m", want: "600000ms"},
		{input: "1h30m", want: "5400000ms"},
		{input: "250ms", want: "250ms"},
		{input: "10min", want: "10min"},
		{input: "0", want: "0ms"},
	}

	for _, test := range tests {
		got, err := normalizeTimeout(test.input)
		if err != nil {
			t.Fatalf("normalizeTimeout(%q) error = %v", test.input, err)
		}
		if got != test.want {
			t.Fatalf("normalizeTimeout(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
