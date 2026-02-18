package hash

import "testing"

func TestCanonicalJSONDeterministic(t *testing.T) {
	a := map[string]any{"b": 2, "a": 1}
	b := map[string]any{"a": 1, "b": 2}
	ha, _, err := HashCanonicalJSON(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, _, err := HashCanonicalJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha != hb {
		t.Fatalf("expected equal digests, got %s vs %s", ha, hb)
	}
}
