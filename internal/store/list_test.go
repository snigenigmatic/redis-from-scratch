package store

import (
	"testing"
)

func TestListPushPopRange(t *testing.T) {
	store := New()

	// LPUSH a b c -> list: c b a
	l, err := store.ListLPush("l1", "a", "b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l != 3 {
		t.Fatalf("expected length 3, got %d", l)
	}

	// LRANGE 0 -1 should return full list: c b a
	arr, err := store.ListRange("l1", 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 3 || arr[0] != "c" || arr[1] != "b" || arr[2] != "a" {
		t.Fatalf("unexpected range result: %v", arr)
	}

	// RPUSH x -> list: c b a x
	l, err = store.ListRPush("l1", "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l != 4 {
		t.Fatalf("expected length 4, got %d", l)
	}

	// LPOP -> c
	v, ok, err := store.ListLPop("l1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || v != "c" {
		t.Fatalf("expected c, got %s ok=%v", v, ok)
	}

	// RPOP -> x
	v, ok, err = store.ListRPop("l1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || v != "x" {
		t.Fatalf("expected x, got %s ok=%v", v, ok)
	}
}

func TestListWrongType(t *testing.T) {
	store := New()
	store.Set("k1", "s", 0)
	_, err := store.ListLPush("k1", "a")
	if err == nil {
		t.Fatalf("expected error when LPUSH on string key")
	}
	_, err = store.ListRPush("k1", "a")
	if err == nil {
		t.Fatalf("expected error when RPUSH on string key")
	}
	_, _, err = store.ListLPop("k1")
	if err == nil {
		t.Fatalf("expected error when LPOP on string key")
	}
}
