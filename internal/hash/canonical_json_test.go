package hash

import (
	"encoding/json"
	"strings"
	"testing"
)

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

func TestCanonicalJSONComplexTypes(t *testing.T) {
	input := map[string]any{
		"z": []any{true, nil, json.Number("3.14")},
		"a": map[string]any{
			"nested": "value",
		},
	}
	canonical, err := CanonicalJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	got := string(canonical)
	if !strings.HasPrefix(got, `{"a":{"nested":"value"},"z":[true,null,3.14]}`) {
		t.Fatalf("unexpected canonical output: %s", got)
	}
}

func TestCanonicalJSONMarshalError(t *testing.T) {
	_, err := CanonicalJSON(map[string]any{"bad": make(chan int)})
	if err == nil {
		t.Fatal("expected marshal error")
	}
	if !strings.Contains(err.Error(), "marshal for canonicalization") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteNumberInvalid(t *testing.T) {
	err := writeNumber(&strings.Builder{}, "not-a-number")
	if err == nil {
		t.Fatal("expected invalid number error")
	}
}
