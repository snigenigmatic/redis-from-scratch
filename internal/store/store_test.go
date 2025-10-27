// tests for internal/store/store.go
package store

import (
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	store := New()
	store.Set("key1", "value1", 0)
	val, ok := store.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected value1, got %s", val)
	}
}

func TestSetGetWithExpiry(t *testing.T) {
	store := New()
	store.Set("key2", "value2", 100) // 100 ms expiry
	val, ok := store.Get("key2")
	if !ok || val != "value2" {
		t.Errorf("Expected value2, got %s", val)
	}

	time.Sleep(150 * time.Millisecond)
	_, ok = store.Get("key2")
	if ok {
		t.Errorf("Expected key2 to be expired")
	}
}

func TestDelete(t *testing.T) {
	store := New()
	store.Set("key3", "value3", 0)
	deleted := store.Delete("key3")
	if deleted != 1 {
		t.Errorf("Expected 1 key to be deleted, got %d", deleted)
	}

	_, ok := store.Get("key3")
	if ok {
		t.Errorf("Expected key3 to be deleted")
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	store := New()
	deleted := store.Delete("nonexistent")
	if deleted != 0 {
		t.Errorf("Expected 0 keys to be deleted, got %d", deleted)
	}
}

func TestExists(t *testing.T) {
	store := New()
	store.Set("key4", "value4", 0)
	count := store.Exists("key4", "nonexistent")
	if count != 1 {
		t.Errorf("Expected 1 existing key, got %d", count)
	}
}
