package sanitize

import "testing"

func TestNameReplacesInvalidCharacters(t *testing.T) {
	got := Name(`a<b>c:d"e/f\g|h?i*j`)
	want := "a_b_c_d_e_f_g_h_i_j"
	if got != want {
		t.Fatalf("Name mismatch: got %q want %q", got, want)
	}
}

func TestNameTrimsUnsafeTrailingCharacters(t *testing.T) {
	got := Name("  Bangumi.  ")
	if got != "Bangumi" {
		t.Fatalf("Name mismatch: got %q", got)
	}
}

func TestNameFallback(t *testing.T) {
	got := Name(`<>:"/\|?*   ...`)
	if got != "untitled" {
		t.Fatalf("Name mismatch: got %q", got)
	}
}
