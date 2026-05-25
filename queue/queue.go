// Package queue provides a generic FIFO (first-in, first-out) queue.
package queue

import "fmt"

const minCapacity = 8

// Queue is a generic FIFO container backed by a growable ring buffer.
//
// Using a ring buffer avoids the memory leak that occurs with a naive
// slice-based queue: when elements are dequeued from the front of a plain
// slice (data = data[1:]), the underlying array retains references to
// popped elements, preventing garbage collection. A ring buffer reuses
// slots and explicitly zeroes dequeued positions.
//
// The zero value is NOT ready for use; create instances with [New]
// or [NewFrom].
//
// Time complexities:
//   - Enqueue: O(1) amortized
//   - Dequeue: O(1) amortized
//   - Peek:    O(1)
//
// Queue is not safe for concurrent use. If concurrent access is required,
// callers must synchronize externally.
type Queue[T any] struct {
	buf   []T
	head  int // index of the front element
	tail  int // index of the next write position
	count int // number of elements
}

// New creates a new empty Queue.
//
// Example:
//
//	q := queue.New[int]()
//	q.Enqueue(1)
//	q.Enqueue(2)
//	fmt.Println(q.Dequeue()) // 1
func New[T any]() *Queue[T] {
	return &Queue[T]{
		buf: make([]T, minCapacity),
	}
}

// NewWithCapacity creates a new empty Queue with pre-allocated capacity.
//
// The actual buffer size will be rounded up to the next power of two
// (minimum 8) to enable efficient modular arithmetic.
func NewWithCapacity[T any](capacity int) *Queue[T] {
	cap := minCapacity
	for cap < capacity {
		cap <<= 1
	}
	return &Queue[T]{
		buf: make([]T, cap),
	}
}

// NewFrom creates a new Queue from an existing slice.
//
// Elements are enqueued in slice order: items[0] will be at the front
// of the queue. The input slice is copied; modifications to the original
// slice after construction will not affect the queue.
func NewFrom[T any](items []T) *Queue[T] {
	cap := minCapacity
	for cap < len(items) {
		cap <<= 1
	}

	buf := make([]T, cap)
	copy(buf, items)

	return &Queue[T]{
		buf:   buf,
		head:  0,
		tail:  len(items) & (cap - 1),
		count: len(items),
	}
}

// Enqueue adds an element to the back of the queue.
//
// Time complexity: O(1) amortized.
func (q *Queue[T]) Enqueue(value T) {
	if q.count == len(q.buf) {
		q.grow()
	}

	q.buf[q.tail] = value
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// Dequeue removes and returns the front element of the queue.
//
// Dequeue panics if the queue is empty. Use [Queue.Len] or [Queue.IsEmpty]
// to check before calling.
//
// Time complexity: O(1) amortized.
func (q *Queue[T]) Dequeue() T {
	if q.count == 0 {
		panic("queue: Dequeue called on empty queue")
	}

	value := q.buf[q.head]

	// Zero the slot to allow garbage collection of referenced objects.
	var zero T
	q.buf[q.head] = zero

	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--

	// Shrink if the buffer is oversized (at most 1/4 full, above minimum).
	if len(q.buf) > minCapacity && q.count > 0 && q.count <= len(q.buf)/4 {
		q.shrink()
	}

	return value
}

// TryDequeue removes and returns the front element of the queue.
//
// Unlike [Queue.Dequeue], it returns a boolean indicating success instead of
// panicking when the queue is empty.
func (q *Queue[T]) TryDequeue() (T, bool) {
	if q.count == 0 {
		var zero T
		return zero, false
	}
	return q.Dequeue(), true
}

// Peek returns the front element without removing it.
//
// Peek panics if the queue is empty. Use [Queue.Len] or [Queue.IsEmpty]
// to check before calling.
//
// Time complexity: O(1).
func (q *Queue[T]) Peek() T {
	if q.count == 0 {
		panic("queue: Peek called on empty queue")
	}
	return q.buf[q.head]
}

// TryPeek returns the front element without removing it.
//
// Unlike [Queue.Peek], it returns a boolean indicating success instead of
// panicking when the queue is empty.
func (q *Queue[T]) TryPeek() (T, bool) {
	if q.count == 0 {
		var zero T
		return zero, false
	}
	return q.buf[q.head], true
}

// PeekBack returns the back (most recently enqueued) element without
// removing it.
//
// PeekBack panics if the queue is empty.
//
// Time complexity: O(1).
func (q *Queue[T]) PeekBack() T {
	if q.count == 0 {
		panic("queue: PeekBack called on empty queue")
	}
	return q.buf[(q.tail-1)&(len(q.buf)-1)]
}

// TryPeekBack returns the back element without removing it.
//
// Unlike [Queue.PeekBack], it returns a boolean indicating success instead
// of panicking when the queue is empty.
func (q *Queue[T]) TryPeekBack() (T, bool) {
	if q.count == 0 {
		var zero T
		return zero, false
	}
	return q.buf[(q.tail-1)&(len(q.buf)-1)], true
}

// Len returns the number of elements in the queue.
func (q *Queue[T]) Len() int {
	return q.count
}

// IsEmpty reports whether the queue has no elements.
func (q *Queue[T]) IsEmpty() bool {
	return q.count == 0
}

// Clear removes all elements from the queue, releasing the underlying storage.
func (q *Queue[T]) Clear() {
	// Zero all live slots to release references for GC.
	var zero T
	for i := 0; i < q.count; i++ {
		idx := (q.head + i) & (len(q.buf) - 1)
		q.buf[idx] = zero
	}

	q.buf = make([]T, minCapacity)
	q.head = 0
	q.tail = 0
	q.count = 0
}

// Values returns a copy of all elements in queue order (front to back).
//
// The returned slice is independent of the queue; modifying it will not
// affect the queue's contents.
func (q *Queue[T]) Values() []T {
	out := make([]T, q.count)
	for i := 0; i < q.count; i++ {
		out[i] = q.buf[(q.head+i)&(len(q.buf)-1)]
	}
	return out
}

// String returns a human-readable representation of the queue's contents.
//
// Elements are listed from front to back. This is intended for debugging
// and logging; the format may change between versions.
func (q *Queue[T]) String() string {
	return fmt.Sprintf("Queue%v", q.Values())
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// grow doubles the buffer capacity and linearises the ring.
func (q *Queue[T]) grow() {
	newBuf := make([]T, len(q.buf)<<1)
	q.copyTo(newBuf)
	q.buf = newBuf
	q.head = 0
	q.tail = q.count
}

// shrink halves the buffer capacity and linearises the ring.
func (q *Queue[T]) shrink() {
	newCap := len(q.buf) >> 1
	if newCap < minCapacity {
		newCap = minCapacity
	}

	newBuf := make([]T, newCap)
	q.copyTo(newBuf)
	q.buf = newBuf
	q.head = 0
	q.tail = q.count
}

// copyTo copies the queue's live elements (in order) into dst.
func (q *Queue[T]) copyTo(dst []T) {
	if q.head < q.tail {
		// Contiguous region: [head, tail).
		copy(dst, q.buf[q.head:q.tail])
	} else if q.count > 0 {
		// Wrapped: [head, end) + [0, tail).
		n := copy(dst, q.buf[q.head:])
		copy(dst[n:], q.buf[:q.tail])
	}
}
