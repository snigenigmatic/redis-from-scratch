package store

import (
	"reflect"
	"testing"
)

func TestZAddAndZRange(t *testing.T) {
	s := New()

	// Add two members
	n, err := s.ZAdd("myz", 1.0, "a")
	if err != nil {
		t.Fatalf("unexpected error adding a: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 added, got %d", n)
	}

	n, err = s.ZAdd("myz", 2.0, "b")
	if err != nil {
		t.Fatalf("unexpected error adding b: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 added for b, got %d", n)
	}

	// Update score for member a (should return 1 for an update to a different score)
	n, err = s.ZAdd("myz", 3.0, "a")
	if err != nil {
		t.Fatalf("unexpected error updating a: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 returned for update, got %d", n)
	}

	// ZRange 0 -1 should return members ordered by score: b(2.0), a(3.0)
	got, err := s.ZRange("myz", 0, -1)
	if err != nil {
		t.Fatalf("unexpected error on ZRange: %v", err)
	}
	want := []string{"b", "a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ZRange returned %v, want %v", got, want)
	}

	// Negative index: ZRange 1 -1 should return [a]
	got, err = s.ZRange("myz", 1, -1)
	if err != nil {
		t.Fatalf("unexpected error on ZRange with negative index: %v", err)
	}
	want = []string{"a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ZRange returned %v, want %v", got, want)
	}

	// ZRange on non-existing key should return empty slice
	got, err = s.ZRange("nope", 0, -1)
	if err != nil {
		t.Fatalf("unexpected error on ZRange for missing key: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result for missing key, got %v", got)
	}

	// Wrong type: create a string key and ensure ZAdd returns WRONGTYPE
	s.Set("strkey", "val", 0)
	_, err = s.ZAdd("strkey", 1.0, "m")
	if err == nil {
		t.Fatalf("expected error when ZAdd on non-zset key")
	}
}
