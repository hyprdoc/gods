package stack

import (
	"strings"
	"testing"
)

func TestPushPop(t *testing.T) {
	s := New[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	if got := s.Pop(); got != 3 {
		t.Errorf("Pop() = %d, want 3", got)
	}
	if got := s.Pop(); got != 2 {
		t.Errorf("Pop() = %d, want 2", got)
	}
	if got := s.Pop(); got != 1 {
		t.Errorf("Pop() = %d, want 1", got)
	}
	if !s.IsEmpty() {
		t.Error("expected empty stack")
	}
}

func TestPeek(t *testing.T) {
	s := New[string]()
	s.Push("a")
	s.Push("b")

	if got := s.Peek(); got != "b" {
		t.Errorf("Peek() = %q, want \"b\"", got)
	}
	if s.Len() != 2 {
		t.Errorf("Peek should not remove element, Len() = %d", s.Len())
	}
}

func TestTryPopEmpty(t *testing.T) {
	s := New[int]()
	_, ok := s.TryPop()
	if ok {
		t.Error("TryPop on empty stack should return false")
	}
}

func TestTryPeekEmpty(t *testing.T) {
	s := New[int]()
	_, ok := s.TryPeek()
	if ok {
		t.Error("TryPeek on empty stack should return false")
	}
}

func TestPopPanicsOnEmpty(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "Pop") {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	New[int]().Pop()
}

func TestPeekPanicsOnEmpty(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "Peek") {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	New[int]().Peek()
}

func TestNewFrom(t *testing.T) {
	s := NewFrom([]int{10, 20, 30})

	if s.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", s.Len())
	}
	if got := s.Pop(); got != 30 {
		t.Errorf("top should be last element of input slice, got %d", got)
	}
}

func TestClear(t *testing.T) {
	s := NewFrom([]int{1, 2, 3})
	s.Clear()

	if !s.IsEmpty() {
		t.Error("expected empty after Clear")
	}
	// Usable after clear.
	s.Push(42)
	if got := s.Peek(); got != 42 {
		t.Errorf("Peek() = %d, want 42", got)
	}
}
