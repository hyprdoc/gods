// Package stack provides a generic LIFO (last-in, first-out) stack.
package stack

import "fmt"

// Stack is a generic LIFO container backed by a slice.
//
// Elements are pushed onto and popped from the top of the stack.
// The zero value is NOT ready for use; create instances with [New]
// or [NewFrom].
//
// Time complexities:
//   - Push: O(1) amortized
//   - Pop:  O(1)
//   - Peek: O(1)
//
// Stack is not safe for concurrent use. If concurrent access is required,
// callers must synchronize externally.
type Stack[T any] struct {
	data []T
}

// New creates a new empty Stack.
//
// Example:
//
//	s := stack.New[int]()
//	s.Push(1)
//	s.Push(2)
//	fmt.Println(s.Pop()) // 2
func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

// NewWithCapacity creates a new empty Stack with pre-allocated capacity.
//
// This is useful when the approximate maximum size is known in advance,
// avoiding repeated slice growth.
func NewWithCapacity[T any](capacity int) *Stack[T] {
	return &Stack[T]{
		data: make([]T, 0, capacity),
	}
}

// NewFrom creates a new Stack from an existing slice.
//
// The last element of the slice becomes the top of the stack.
// The input slice is copied; modifications to the original slice after
// construction will not affect the stack.
func NewFrom[T any](items []T) *Stack[T] {
	data := make([]T, len(items))
	copy(data, items)
	return &Stack[T]{data: data}
}

// Push adds an element to the top of the stack.
//
// Time complexity: O(1) amortized.
func (s *Stack[T]) Push(value T) {
	s.data = append(s.data, value)
}

// Pop removes and returns the top element of the stack.
//
// Pop panics if the stack is empty. Use [Stack.Len] or [Stack.IsEmpty]
// to check before calling.
//
// Time complexity: O(1).
func (s *Stack[T]) Pop() T {
	if len(s.data) == 0 {
		panic("stack: Pop called on empty stack")
	}

	top := len(s.data) - 1
	value := s.data[top]

	// Zero the slot to allow garbage collection of referenced objects.
	var zero T
	s.data[top] = zero
	s.data = s.data[:top]

	return value
}

// TryPop removes and returns the top element of the stack.
//
// Unlike [Stack.Pop], it returns a boolean indicating success instead of
// panicking when the stack is empty.
func (s *Stack[T]) TryPop() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.Pop(), true
}

// Peek returns the top element without removing it.
//
// Peek panics if the stack is empty. Use [Stack.Len] or [Stack.IsEmpty]
// to check before calling.
//
// Time complexity: O(1).
func (s *Stack[T]) Peek() T {
	if len(s.data) == 0 {
		panic("stack: Peek called on empty stack")
	}
	return s.data[len(s.data)-1]
}

// TryPeek returns the top element without removing it.
//
// Unlike [Stack.Peek], it returns a boolean indicating success instead of
// panicking when the stack is empty.
func (s *Stack[T]) TryPeek() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Len returns the number of elements in the stack.
func (s *Stack[T]) Len() int {
	return len(s.data)
}

// IsEmpty reports whether the stack has no elements.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// Clear removes all elements from the stack, releasing the underlying storage.
func (s *Stack[T]) Clear() {
	s.data = nil
}

// Values returns a copy of all elements from bottom to top.
//
// The returned slice is independent of the stack; modifying it will not
// affect the stack's contents.
func (s *Stack[T]) Values() []T {
	out := make([]T, len(s.data))
	copy(out, s.data)
	return out
}

// String returns a human-readable representation of the stack's contents.
//
// Elements are listed from bottom to top. This is intended for debugging
// and logging; the format may change between versions.
func (s *Stack[T]) String() string {
	return fmt.Sprintf("Stack%v", s.data)
}
