package store

import (
	"fmt"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	s := New()
	defer s.Shutdown()

	s.Set("name", "aditya", 0)
	val, ok := s.Get("name")
	if !ok || val != "aditya" {
		t.Fatalf("expected 'aditya', got '%s' (ok=%v)", val, ok)
	}
}

func TestGetMissing(t *testing.T) {
	s := New()
	defer s.Shutdown()

	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing key")
	}
}

func TestDel(t *testing.T) {
	s := New()
	defer s.Shutdown()

	s.Set("a", "1", 0)
	s.Set("b", "2", 0)
	s.Set("c", "3", 0)

	if n := s.Del("a", "b"); n != 2 {
		t.Fatalf("expected 2 deleted, got %d", n)
	}
	if n := s.Del("nonexistent"); n != 0 {
		t.Fatalf("expected 0 for nonexistent, got %d", n)
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("expected 'a' to be deleted")
	}
	if _, ok := s.Get("c"); !ok {
		t.Fatal("expected 'c' to still exist")
	}
}

func TestTTL(t *testing.T) {
	s := New()
	defer s.Shutdown()

	// Key without TTL
	s.Set("perm", "forever", 0)
	if ttl := s.TTL("perm"); ttl != -1 {
		t.Fatalf("expected -1 for no TTL, got %d", ttl)
	}

	// Key with TTL
	s.Set("temp", "gone", 2*time.Second)
	ttl := s.TTL("temp")
	if ttl <= 0 || ttl > 3 {
		t.Fatalf("expected positive TTL, got %d", ttl)
	}

	// Wait for expiry
	s.Set("quick", "flash", 50*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	if ttl := s.TTL("quick"); ttl != -2 {
		t.Fatalf("expected -2 for expired key, got %d", ttl)
	}
}

func TestExpire(t *testing.T) {
	s := New()
	defer s.Shutdown()

	s.Set("key", "val", 0)
	if ok := s.Expire("key", 2*time.Second); !ok {
		t.Fatal("expected Expire to return true")
	}
	if ttl := s.TTL("key"); ttl <= 0 {
		t.Fatalf("expected positive TTL, got %d", ttl)
	}

	// Expire on a short-lived key and wait for expiry
	s.Set("short", "lived", 0)
	s.Expire("short", 50*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	if _, ok := s.Get("short"); ok {
		t.Fatal("expected key to be expired")
	}

	// Expire on nonexistent key
	if ok := s.Expire("nonexistent", 10*time.Second); ok {
		t.Fatal("expected false for nonexistent key")
	}
}

func TestExists(t *testing.T) {
	s := New()
	defer s.Shutdown()

	s.Set("a", "1", 0)
	if n := s.Exists("a"); n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
	if n := s.Exists("b"); n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestKeys(t *testing.T) {
	s := New()
	defer s.Shutdown()

	s.Set("a", "1", 0)
	s.Set("b", "2", 0)
	s.Set("c", "3", 0)

	keys := s.Keys("*")
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d: %v", len(keys), keys)
	}
}

func TestFlush(t *testing.T) {
	s := New()
	defer s.Shutdown()

	s.Set("a", "1", 0)
	s.Set("b", "2", 0)
	s.Flush()

	if n := s.Len(); n != 0 {
		t.Fatalf("expected 0 after flush, got %d", n)
	}
}

func TestConcurrency(t *testing.T) {
	s := New()
	defer s.Shutdown()

	done := make(chan struct{})
	n := 100

	for i := 0; i < n; i++ {
		go func(i int) {
			s.Set(fmt.Sprintf("key-%d", i), fmt.Sprintf("val-%d", i), 0)
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < n; i++ {
		<-done
	}

	if l := s.Len(); l != n {
		t.Fatalf("expected %d keys, got %d", n, l)
	}
}

func TestStats(t *testing.T) {
	s := New()
	defer s.Shutdown()

	s.Set("a", "1", 0)
	stats := s.Stats()
	if stats["keys"].(int) != 1 {
		t.Fatalf("expected 1 key, got %d", stats["keys"])
	}
}