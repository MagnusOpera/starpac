package archive

import "testing"

func TestValidateNameRejectsTraversal(t *testing.T) {
	for _, name := range []string{"../secret", "scripts/../../secret", "/absolute"} {
		if err := ValidateName(name); err == nil {
			t.Fatalf("ValidateName(%q) accepted an unsafe path", name)
		}
	}
	if err := ValidateName("scripts/Tables/widgets.sql"); err != nil {
		t.Fatalf("ValidateName rejected package entry: %v", err)
	}
}

func TestChecksumIsStable(t *testing.T) {
	const expected = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := Checksum([]byte("abc")); got != expected {
		t.Fatalf("Checksum() = %q, want %q", got, expected)
	}
}
