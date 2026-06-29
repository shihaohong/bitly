package links

import (
	"testing"
	"unicode"
)

func TestGenerateCode(t *testing.T) {
	seen := make(map[string]bool, 100)

	for i := 0; i < 100; i++ {
		code, err := generateCode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(code) != codeLen {
			t.Errorf("code %q: len=%d, want %d", code, len(code), codeLen)
		}

		for _, ch := range code {
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
				t.Errorf("code %q contains non-alphanumeric char %q", code, ch)
			}
		}

		seen[code] = true
	}

	// Collisions from a 62^7 space over 100 draws are statistically negligible.
	if len(seen) < 95 {
		t.Errorf("too many collisions: %d unique codes out of 100", len(seen))
	}
}
