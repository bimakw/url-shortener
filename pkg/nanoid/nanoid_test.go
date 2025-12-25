package nanoid

import (
	"regexp"
	"testing"
)

func TestGenerate(t *testing.T) {
	id, err := Generate(8)
	if err != nil {
		t.Fatalf("Generate(8) returned error: %v", err)
	}
	if len(id) != 8 {
		t.Errorf("Generate(8) returned id with length %d, want 8", len(id))
	}
}

func TestGenerateUsesAlphabet(t *testing.T) {
	// Generate multiple IDs and verify all characters are from alphabet
	pattern := regexp.MustCompile("^[0-9A-Za-z]+$")

	for i := 0; i < 100; i++ {
		id, err := Generate(DefaultSize)
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if !pattern.MatchString(id) {
			t.Errorf("Generated ID %q contains characters not in alphabet", id)
		}
	}
}

func TestGenerateWithZeroSize(t *testing.T) {
	id, err := Generate(0)
	if err != nil {
		t.Fatalf("Generate(0) returned error: %v", err)
	}
	if len(id) != DefaultSize {
		t.Errorf("Generate(0) returned id with length %d, want %d (DefaultSize)", len(id), DefaultSize)
	}
}

func TestGenerateWithNegativeSize(t *testing.T) {
	id, err := Generate(-5)
	if err != nil {
		t.Fatalf("Generate(-5) returned error: %v", err)
	}
	if len(id) != DefaultSize {
		t.Errorf("Generate(-5) returned id with length %d, want %d (DefaultSize)", len(id), DefaultSize)
	}
}

func TestGenerateWithLargeSize(t *testing.T) {
	size := 32
	id, err := Generate(size)
	if err != nil {
		t.Fatalf("Generate(%d) returned error: %v", size, err)
	}
	if len(id) != size {
		t.Errorf("Generate(%d) returned id with length %d", size, len(id))
	}
}

func TestGenerateUniqueness(t *testing.T) {
	// Generate many IDs and check for collisions
	ids := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		id, err := Generate(DefaultSize)
		if err != nil {
			t.Fatalf("Generate returned error: %v", err)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}

func TestNew(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if len(id) != DefaultSize {
		t.Errorf("New() returned id with length %d, want %d", len(id), DefaultSize)
	}
}

func TestMustGenerate(t *testing.T) {
	// Should not panic
	id := MustGenerate(8)
	if len(id) != 8 {
		t.Errorf("MustGenerate(8) returned id with length %d, want 8", len(id))
	}
}

func TestMustGenerateWithDefaultSize(t *testing.T) {
	id := MustGenerate(0)
	if len(id) != DefaultSize {
		t.Errorf("MustGenerate(0) returned id with length %d, want %d", len(id), DefaultSize)
	}
}

func TestDefaultSize(t *testing.T) {
	if DefaultSize != 8 {
		t.Errorf("DefaultSize = %d, want 8", DefaultSize)
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Generate(DefaultSize)
	}
}
