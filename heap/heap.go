package heap

import (
	"cmp"
	"fmt"
)

// Heap is a generic binary heap backed by a slice.
//
// A Heap maintains the heap invariant: the element at the root is always the
// "minimum" according to the configured comparison function. By supplying
// different comparison functions, callers can create min-heaps, max-heaps,
// or heaps ordered by any arbitrary criteria.
//
// The zero value is NOT ready for use; create instances with [New],
// [NewMin], or [NewMax].
//
// Time complexities:
//   - Push:    O(log n)
//   - Pop:     O(log n)
//   - Peek:    O(1)
//   - Remove:  O(n)
//   - Contains: O(n)
//
// Heap is not safe for concurrent use. If concurrent access is required,
// callers must synchronize externally.
type Heap[T any] struct {
	data []T
	less func(a, b T) bool
}

// New creates a new empty Heap ordered by the provided comparison function.
//
// The less function defines the heap ordering. It should return true when a
// should be closer to the root of the heap than b.
//
// For a min-heap, less should return a < b.
// For a max-heap, less should return a > b.
//
// Example:
//
//	// Custom struct heap ordered by priority.
//	h := heap.New(func(a, b Task) bool {
//	    return a.Priority < b.Priority
//	})
func New[T any](less func(a, b T) bool) *Heap[T] {
	return &Heap[T]{
		less: less,
	}
}

// NewMin creates a new empty min-heap for any [cmp.Ordered] type.
//
// The smallest element will be at the root (returned first by [Heap.Pop]).
//
// Example:
//
//	h := heap.NewMin[int]()
//	h.Push(3)
//	h.Push(1)
//	h.Push(2)
//	fmt.Println(h.Pop()) // 1
func NewMin[T cmp.Ordered]() *Heap[T] {
	return &Heap[T]{
		less: func(a, b T) bool { return cmp.Less(a, b) },
	}
}

// NewMax creates a new empty max-heap for any [cmp.Ordered] type.
//
// The largest element will be at the root (returned first by [Heap.Pop]).
//
// Example:
//
//	h := heap.NewMax[int]()
//	h.Push(3)
//	h.Push(1)
//	h.Push(2)
//	fmt.Println(h.Pop()) // 3
func NewMax[T cmp.Ordered]() *Heap[T] {
	return &Heap[T]{
		less: func(a, b T) bool { return cmp.Less(b, a) },
	}
}

// NewFrom creates a new Heap from an existing slice, ordered by the
// provided comparison function.
//
// This is more efficient than creating an empty heap and pushing elements
// one by one: it runs in O(n) time via Floyd's heap-construction algorithm,
// compared to O(n log n) for repeated pushes.
//
// The input slice is copied; modifications to the original slice after
// construction will not affect the heap.
func NewFrom[T any](items []T, less func(a, b T) bool) *Heap[T] {
	data := make([]T, len(items))
	copy(data, items)

	h := &Heap[T]{
		data: data,
		less: less,
	}
	h.heapify()
	return h
}

// NewMinFrom creates a new min-heap from an existing slice of
// [cmp.Ordered] elements in O(n) time.
func NewMinFrom[T cmp.Ordered](items []T) *Heap[T] {
	return NewFrom(items, func(a, b T) bool { return cmp.Less(a, b) })
}

// NewMaxFrom creates a new max-heap from an existing slice of
// [cmp.Ordered] elements in O(n) time.
func NewMaxFrom[T cmp.Ordered](items []T) *Heap[T] {
	return NewFrom(items, func(a, b T) bool { return cmp.Less(b, a) })
}

// Push adds an element to the heap, maintaining the heap invariant.
//
// Time complexity: O(log n).
func (h *Heap[T]) Push(value T) {
	h.data = append(h.data, value)
	h.siftUp(len(h.data) - 1)
}

// Pop removes and returns the root element of the heap (the minimum element
// according to the comparison function).
//
// Pop panics if the heap is empty. Use [Heap.Len] or [Heap.IsEmpty] to check
// before calling.
//
// Time complexity: O(log n).
func (h *Heap[T]) Pop() T {
	if len(h.data) == 0 {
		panic("heap: Pop called on empty heap")
	}

	root := h.data[0]
	last := len(h.data) - 1

	h.data[0] = h.data[last]

	// Clear the last element to avoid retaining references in the
	// underlying array, which could prevent garbage collection.
	var zero T
	h.data[last] = zero
	h.data = h.data[:last]

	if len(h.data) > 0 {
		h.siftDown(0)
	}

	return root
}

// TryPop removes and returns the root element of the heap.
//
// Unlike [Heap.Pop], it returns a boolean indicating success instead of
// panicking when the heap is empty.
func (h *Heap[T]) TryPop() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	return h.Pop(), true
}

// Peek returns the root element without removing it.
//
// Peek panics if the heap is empty. Use [Heap.Len] or [Heap.IsEmpty] to
// check before calling.
//
// Time complexity: O(1).
func (h *Heap[T]) Peek() T {
	if len(h.data) == 0 {
		panic("heap: Peek called on empty heap")
	}
	return h.data[0]
}

// TryPeek returns the root element without removing it.
//
// Unlike [Heap.Peek], it returns a boolean indicating success instead of
// panicking when the heap is empty.
func (h *Heap[T]) TryPeek() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	return h.data[0], true
}

// PushPop pushes value onto the heap and then pops the root element,
// which is more efficient than calling Push followed by Pop separately.
//
// This is useful for maintaining a fixed-size heap (e.g., a top-K tracker).
//
// Time complexity: O(log n).
func (h *Heap[T]) PushPop(value T) T {
	if len(h.data) > 0 && h.less(h.data[0], value) {
		value, h.data[0] = h.data[0], value
		h.siftDown(0)
	}
	return value
}

// Replace pops the root element and then pushes value, which is more
// efficient than calling Pop followed by Push separately.
//
// Replace panics if the heap is empty.
//
// Time complexity: O(log n).
func (h *Heap[T]) Replace(value T) T {
	if len(h.data) == 0 {
		panic("heap: Replace called on empty heap")
	}

	root := h.data[0]
	h.data[0] = value
	h.siftDown(0)
	return root
}

// Remove removes the first occurrence of value from the heap, using the
// provided equal function to test for equality.
//
// Returns true if the value was found and removed, false otherwise.
//
// Time complexity: O(n) for the search, O(log n) for the removal.
func (h *Heap[T]) Remove(equal func(T) bool) bool {
	for i, v := range h.data {
		if equal(v) {
			h.removeAt(i)
			return true
		}
	}
	return false
}

// Contains reports whether the heap contains an element satisfying the
// provided predicate.
//
// Time complexity: O(n).
func (h *Heap[T]) Contains(predicate func(T) bool) bool {
	for _, v := range h.data {
		if predicate(v) {
			return true
		}
	}
	return false
}

// Len returns the number of elements in the heap.
func (h *Heap[T]) Len() int {
	return len(h.data)
}

// IsEmpty reports whether the heap has no elements.
func (h *Heap[T]) IsEmpty() bool {
	return len(h.data) == 0
}

// Clear removes all elements from the heap, releasing the underlying storage.
func (h *Heap[T]) Clear() {
	h.data = nil
}

// Values returns a copy of all elements in the heap in no particular order.
//
// The returned slice is independent of the heap; modifying it will not
// affect the heap's contents.
func (h *Heap[T]) Values() []T {
	out := make([]T, len(h.data))
	copy(out, h.data)
	return out
}

// Sorted returns a copy of all elements in heap order (root first).
//
// This is equivalent to repeatedly calling [Heap.Pop] on a copy of the heap.
// The original heap is not modified.
//
// Time complexity: O(n log n).
func (h *Heap[T]) Sorted() []T {
	// Work on a copy so the original heap is unchanged.
	clone := &Heap[T]{
		data: make([]T, len(h.data)),
		less: h.less,
	}
	copy(clone.data, h.data)

	result := make([]T, 0, len(h.data))
	for clone.Len() > 0 {
		result = append(result, clone.Pop())
	}
	return result
}

// String returns a human-readable representation of the heap's contents.
//
// This is intended for debugging and logging; the format may change between
// versions and should not be parsed programmatically.
func (h *Heap[T]) String() string {
	return fmt.Sprintf("Heap%v", h.data)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// heapify establishes the heap invariant in O(n) time using Floyd's
// bottom-up heap construction algorithm. It starts from the last non-leaf
// node and sifts each node down.
func (h *Heap[T]) heapify() {
	n := len(h.data)
	// Start from the last parent node and sift down each one.
	for i := (n / 2) - 1; i >= 0; i-- {
		h.siftDown(i)
	}
}

// siftUp moves the element at index i upward until the heap invariant is
// restored. Used after inserting a new element at the end of the slice.
func (h *Heap[T]) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(h.data[i], h.data[parent]) {
			break
		}
		h.data[i], h.data[parent] = h.data[parent], h.data[i]
		i = parent
	}
}

// siftDown moves the element at index i downward until the heap invariant
// is restored. Used after removing the root or during heapification.
func (h *Heap[T]) siftDown(i int) {
	n := len(h.data)
	for {
		smallest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < n && h.less(h.data[left], h.data[smallest]) {
			smallest = left
		}
		if right < n && h.less(h.data[right], h.data[smallest]) {
			smallest = right
		}

		if smallest == i {
			break
		}

		h.data[i], h.data[smallest] = h.data[smallest], h.data[i]
		i = smallest
	}
}

// removeAt removes the element at index i, maintaining the heap invariant.
func (h *Heap[T]) removeAt(i int) {
	last := len(h.data) - 1

	if i == last {
		// Removing the last element — no re-heaping needed.
		var zero T
		h.data[last] = zero
		h.data = h.data[:last]
		return
	}

	// Move the last element into the removed position.
	h.data[i] = h.data[last]

	var zero T
	h.data[last] = zero
	h.data = h.data[:last]

	// The replacement element may need to move up or down.
	// Try sifting up first; if it didn't move, sift down.
	parent := (i - 1) / 2
	if i > 0 && h.less(h.data[i], h.data[parent]) {
		h.siftUp(i)
	} else {
		h.siftDown(i)
	}
}
