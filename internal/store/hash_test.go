package store

import (
	"testing"
)

func TestHashSetGetDel(t *testing.T) {
	store := New()
	// HSET new field
	n, err := store.HashSet("h1", "f1", "v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 new field, got %d", n)
	}

	// HGET existing
	v, ok, err := store.HashGet("h1", "f1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || v != "v1" {
		t.Fatalf("expected v1, got '%s' ok=%v", v, ok)
	}

	// HSET update field
	n2, err := store.HashSet("h1", "f1", "v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 (updated), got %d", n2)
	}

	v, ok, err = store.HashGet("h1", "f1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || v != "v2" {
		t.Fatalf("expected v2, got '%s' ok=%v", v, ok)
	}

	// HDEL existing
	delCount, err := store.HashDel("h1", "f1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delCount != 1 {
		t.Fatalf("expected 1 deleted, got %d", delCount)
	}

	// HGET after delete
	_, ok, err = store.HashGet("h1", "f1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected not found after delete")
	}
}

func TestHashWrongType(t *testing.T) {
	store := New()
	// Set a string value at key
	store.Set("k1", "s", 0)
	_, err := store.HashSet("k1", "f1", "v1")
	if err == nil {
		t.Fatalf("expected error when HSET on string key")
	}
	_, _, err = store.HashGet("k1", "f1")
	if err == nil {
		t.Fatalf("expected error when HGET on string key")
	}
}
