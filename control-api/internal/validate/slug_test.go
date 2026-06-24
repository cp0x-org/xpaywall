package validate

import "testing"

func TestSlug(t *testing.T) {
	// Slug gates values placed into public proxy paths (/{username}/{slug}/...),
	// so anything outside [A-Za-z0-9_-] or an empty string must be rejected.
	valid := []string{"admin", "my-project", "user_1", "ABC", "a", "Mix_of-9"}
	for _, s := range valid {
		if !Slug(s) {
			t.Errorf("Slug(%q) = false, want true", s)
		}
	}

	invalid := []string{"", "has space", "dot.name", "slash/here", "emoji😀", "plus+one", "tab\tx", "café"}
	for _, s := range invalid {
		if Slug(s) {
			t.Errorf("Slug(%q) = true, want false", s)
		}
	}
}

func TestSanitize(t *testing.T) {
	// Sanitize must always return a valid Slug (or empty), since its output is
	// used directly as an auto-derived username.
	cases := map[string]string{
		"John Doe":     "John-Doe",
		"a.b.c":        "a-b-c",
		"  spaced  ":   "spaced",
		"user@mail":    "user-mail",
		"already_ok-1": "already_ok-1",
		"!!!":          "",
		"café":         "caf",
	}
	for in, want := range cases {
		got := Sanitize(in)
		if got != want {
			t.Errorf("Sanitize(%q) = %q, want %q", in, got, want)
		}
		if got != "" && !Slug(got) {
			t.Errorf("Sanitize(%q) = %q which is not a valid Slug", in, got)
		}
	}
}
