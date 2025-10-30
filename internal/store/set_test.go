package store

import (
	"testing"
)

func TestSetAddMembers(t *testing.T) {
	store := New()

	// Add members
	n, err := store.SetAdd("s1", "a", "b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected 3 added, got %d", n)
	}

	// Add existing member
	n, err = store.SetAdd("s1", "b", "d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 { // only d added
		t.Fatalf("expected 1 added, got %d", n)
	}

	// Members
	m, err := store.SetMembers("s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 4 {
		t.Fatalf("expected 4 members, got %d", len(m))
	}

	// IsMember
	exists, err := store.SetIsMember("s1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatalf("expected a to be a member")
	}

	// Remove
	r, err := store.SetRemove("s1", "a", "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r != 1 {
		t.Fatalf("expected 1 removed, got %d", r)
	}

	// IsMember after remove
	exists, err = store.SetIsMember("s1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("expected a to be removed")
	}
}

func TestSetWrongType(t *testing.T) {
	store := New()
	store.Set("k1", "s", 0)
	_, err := store.SetAdd("k1", "a")
	if err == nil {
		t.Fatalf("expected error when SADD on string key")
	}
	_, err = store.SetRemove("k1", "a")
	if err == nil {
		t.Fatalf("expected error when SREM on string key")
	}
	_, err = store.SetMembers("k1")
	if err == nil {
		t.Fatalf("expected error when SMEMBERS on string key")
	}
}
