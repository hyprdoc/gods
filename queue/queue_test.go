package queue

import (
	"strings"
	"testing"
)

func TestEnqueueDequeue(t *testing.T) {
	q := New[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	if got := q.Dequeue(); got != 1 {
		t.Errorf("Dequeue() = %d, want 1", got)
	}
	if got := q.Dequeue(); got != 2 {
		t.Errorf("Dequeue() = %d, want 2", got)
	}
	if got := q.Dequeue(); got != 3 {
		t.Errorf("Dequeue() = %d, want 3", got)
	}
	if !q.IsEmpty() {
		t.Error("expected empty queue")
	}
}

func TestPeek(t *testing.T) {
	q := New[string]()
	q.Enqueue("first")
	q.Enqueue("second")

	if got := q.Peek(); got != "first" {
		t.Errorf("Peek() = %q, want \"first\"", got)
	}
	if q.Len() != 2 {
		t.Errorf("Peek should not remove element, Len() = %d", q.Len())
	}
}

func TestPeekBack(t *testing.T) {
	q := New[int]()
	q.Enqueue(10)
	q.Enqueue(20)

	if got := q.PeekBack(); got != 20 {
		t.Errorf("PeekBack() = %d, want 20", got)
	}
}

func TestTryDequeueEmpty(t *testing.T) {
	q := New[int]()
	_, ok := q.TryDequeue()
	if ok {
		t.Error("TryDequeue on empty queue should return false")
	}
}

func TestTryPeekEmpty(t *testing.T) {
	q := New[int]()
	_, ok := q.TryPeek()
	if ok {
		t.Error("TryPeek on empty queue should return false")
	}
}

func TestDequeuePanicsOnEmpty(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, "Dequeue") {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	New[int]().Dequeue()
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
	q := NewFrom([]int{10, 20, 30})

	if q.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", q.Len())
	}
	if got := q.Dequeue(); got != 10 {
		t.Errorf("front should be first element of input slice, got %d", got)
	}
}

func TestWrapAround(t *testing.T) {
	// Force the ring buffer to wrap by enqueuing/dequeuing past the initial capacity.
	q := New[int]()
	for i := 0; i < 20; i++ {
		q.Enqueue(i)
	}
	for i := 0; i < 15; i++ {
		q.Dequeue()
	}
	for i := 20; i < 30; i++ {
		q.Enqueue(i)
	}

	// Remaining: 15..29
	for i := 15; i < 30; i++ {
		got := q.Dequeue()
		if got != i {
			t.Fatalf("Dequeue() = %d, want %d", got, i)
		}
	}
	if !q.IsEmpty() {
		t.Error("expected empty queue")
	}
}

func TestClear(t *testing.T) {
	q := NewFrom([]int{1, 2, 3})
	q.Clear()

	if !q.IsEmpty() {
		t.Error("expected empty after Clear")
	}
	// Usable after clear.
	q.Enqueue(42)
	if got := q.Peek(); got != 42 {
		t.Errorf("Peek() = %d, want 42", got)
	}
}
