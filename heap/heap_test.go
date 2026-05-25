package heap

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// assertHeapInvariant verifies the heap property holds for every node:
// each parent is "less than" (closer to root) both of its children.
func assertHeapInvariant[T any](t *testing.T, h *Heap[T]) {
	t.Helper()
	for i := 1; i < len(h.data); i++ {
		parent := (i - 1) / 2
		if h.less(h.data[i], h.data[parent]) {
			t.Errorf("heap invariant violated: data[%d] should not be less than data[%d] (parent)", i, parent)
		}
	}
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	h := New(func(a, b int) bool { return a < b })
	if h == nil {
		t.Fatal("New returned nil")
	}
	if h.Len() != 0 {
		t.Errorf("expected empty heap, got Len() = %d", h.Len())
	}
	if !h.IsEmpty() {
		t.Error("expected IsEmpty() = true")
	}
}

func TestNewMin(t *testing.T) {
	h := NewMin[int]()
	for _, v := range []int{5, 3, 8, 1, 4} {
		h.Push(v)
	}
	assertHeapInvariant(t, h)

	got := h.Pop()
	if got != 1 {
		t.Errorf("NewMin: expected Pop() = 1, got %d", got)
	}
}

func TestNewMax(t *testing.T) {
	h := NewMax[int]()
	for _, v := range []int{5, 3, 8, 1, 4} {
		h.Push(v)
	}
	assertHeapInvariant(t, h)

	got := h.Pop()
	if got != 8 {
		t.Errorf("NewMax: expected Pop() = 8, got %d", got)
	}
}

func TestNewMin_Strings(t *testing.T) {
	h := NewMin[string]()
	h.Push("cherry")
	h.Push("apple")
	h.Push("banana")

	if got := h.Pop(); got != "apple" {
		t.Errorf("expected Pop() = apple, got %s", got)
	}
}

func TestNewMax_Float64(t *testing.T) {
	h := NewMax[float64]()
	h.Push(1.1)
	h.Push(3.3)
	h.Push(2.2)

	if got := h.Pop(); got != 3.3 {
		t.Errorf("expected Pop() = 3.3, got %f", got)
	}
}

// ---------------------------------------------------------------------------
// From-slice constructors
// ---------------------------------------------------------------------------

func TestNewFrom(t *testing.T) {
	items := []int{9, 4, 7, 1, 3, 6, 2, 8, 5}
	h := NewFrom(items, func(a, b int) bool { return a < b })

	assertHeapInvariant(t, h)
	if h.Len() != len(items) {
		t.Errorf("expected Len() = %d, got %d", len(items), h.Len())
	}
}

func TestNewMinFrom(t *testing.T) {
	items := []int{9, 4, 7, 1, 3, 6, 2, 8, 5}
	h := NewMinFrom(items)

	assertHeapInvariant(t, h)

	prev := h.Pop()
	for h.Len() > 0 {
		cur := h.Pop()
		if cur < prev {
			t.Errorf("min-heap order violated: %d came after %d", cur, prev)
		}
		prev = cur
	}
}

func TestNewMaxFrom(t *testing.T) {
	items := []int{9, 4, 7, 1, 3, 6, 2, 8, 5}
	h := NewMaxFrom(items)

	assertHeapInvariant(t, h)

	prev := h.Pop()
	for h.Len() > 0 {
		cur := h.Pop()
		if cur > prev {
			t.Errorf("max-heap order violated: %d came after %d", cur, prev)
		}
		prev = cur
	}
}

func TestNewFrom_DoesNotMutateInput(t *testing.T) {
	items := []int{3, 1, 2}
	original := make([]int, len(items))
	copy(original, items)

	_ = NewMinFrom(items)

	if !slices.Equal(items, original) {
		t.Errorf("NewMinFrom mutated input slice: got %v, want %v", items, original)
	}
}

func TestNewFrom_EmptySlice(t *testing.T) {
	h := NewMinFrom[int](nil)

	if h.Len() != 0 {
		t.Errorf("expected Len() = 0, got %d", h.Len())
	}
	if !h.IsEmpty() {
		t.Error("expected IsEmpty() = true for heap built from nil slice")
	}
}

// ---------------------------------------------------------------------------
// Push / Pop
// ---------------------------------------------------------------------------

func TestPush_MaintainsInvariant(t *testing.T) {
	h := NewMin[int]()
	values := []int{10, 4, 15, 20, 0, 8, 3, 12, 1, 7}
	for _, v := range values {
		h.Push(v)
		assertHeapInvariant(t, h)
	}
}

func TestPop_MinHeapOrder(t *testing.T) {
	h := NewMin[int]()
	values := []int{10, 4, 15, 20, 0, 8, 3, 12, 1, 7}
	for _, v := range values {
		h.Push(v)
	}

	sorted := make([]int, 0, len(values))
	for !h.IsEmpty() {
		sorted = append(sorted, h.Pop())
	}

	if !slices.IsSorted(sorted) {
		t.Errorf("expected sorted output, got %v", sorted)
	}
}

func TestPop_MaxHeapOrder(t *testing.T) {
	h := NewMax[int]()
	values := []int{10, 4, 15, 20, 0, 8, 3, 12, 1, 7}
	for _, v := range values {
		h.Push(v)
	}

	var prev int
	first := true
	for !h.IsEmpty() {
		cur := h.Pop()
		if !first && cur > prev {
			t.Errorf("max-heap order violated: %d followed %d", cur, prev)
		}
		prev = cur
		first = false
	}
}

func TestPop_SingleElement(t *testing.T) {
	h := NewMin[int]()
	h.Push(42)

	got := h.Pop()
	if got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
	if !h.IsEmpty() {
		t.Error("expected empty heap after popping single element")
	}
}

func TestPop_PanicsOnEmpty(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on Pop from empty heap")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "Pop") {
			t.Errorf("unexpected panic value: %v", r)
		}
	}()

	h := NewMin[int]()
	h.Pop()
}

// ---------------------------------------------------------------------------
// TryPop / TryPeek
// ---------------------------------------------------------------------------

func TestTryPop_Empty(t *testing.T) {
	h := NewMin[int]()
	val, ok := h.TryPop()
	if ok {
		t.Error("expected ok = false on empty heap")
	}
	if val != 0 {
		t.Errorf("expected zero value, got %d", val)
	}
}

func TestTryPop_NonEmpty(t *testing.T) {
	h := NewMin[int]()
	h.Push(5)
	h.Push(3)

	val, ok := h.TryPop()
	if !ok {
		t.Error("expected ok = true")
	}
	if val != 3 {
		t.Errorf("expected 3, got %d", val)
	}
}

func TestTryPeek_Empty(t *testing.T) {
	h := NewMax[string]()
	val, ok := h.TryPeek()
	if ok {
		t.Error("expected ok = false on empty heap")
	}
	if val != "" {
		t.Errorf("expected zero value, got %q", val)
	}
}

func TestTryPeek_NonEmpty(t *testing.T) {
	h := NewMax[int]()
	h.Push(1)
	h.Push(9)

	val, ok := h.TryPeek()
	if !ok {
		t.Error("expected ok = true")
	}
	if val != 9 {
		t.Errorf("expected 9, got %d", val)
	}
	// Peek should not remove the element.
	if h.Len() != 2 {
		t.Errorf("expected Len() = 2 after Peek, got %d", h.Len())
	}
}

// ---------------------------------------------------------------------------
// Peek
// ---------------------------------------------------------------------------

func TestPeek(t *testing.T) {
	h := NewMin[int]()
	h.Push(7)
	h.Push(2)
	h.Push(5)

	if got := h.Peek(); got != 2 {
		t.Errorf("expected Peek() = 2, got %d", got)
	}
	// Len must be unchanged.
	if h.Len() != 3 {
		t.Errorf("expected Len() = 3, got %d", h.Len())
	}
}

func TestPeek_PanicsOnEmpty(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on Peek from empty heap")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "Peek") {
			t.Errorf("unexpected panic value: %v", r)
		}
	}()

	h := NewMin[int]()
	h.Peek()
}

// ---------------------------------------------------------------------------
// PushPop
// ---------------------------------------------------------------------------

func TestPushPop_ValueSmallerThanRoot(t *testing.T) {
	h := NewMin[int]()
	h.Push(5)
	h.Push(10)

	// Pushing 1 (smaller than root 5): PushPop should return 1 immediately
	// since pushing then popping would yield the new value.
	got := h.PushPop(1)
	if got != 1 {
		t.Errorf("expected PushPop(1) = 1, got %d", got)
	}
	assertHeapInvariant(t, h)
	if h.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d", h.Len())
	}
}

func TestPushPop_ValueLargerThanRoot(t *testing.T) {
	h := NewMin[int]()
	h.Push(5)
	h.Push(10)

	// Pushing 7 (larger than root 5): should return old root (5) and insert 7.
	got := h.PushPop(7)
	if got != 5 {
		t.Errorf("expected PushPop(7) = 5, got %d", got)
	}
	assertHeapInvariant(t, h)
}

func TestPushPop_EmptyHeap(t *testing.T) {
	h := NewMin[int]()

	// On an empty heap, PushPop should just return the value.
	got := h.PushPop(42)
	if got != 42 {
		t.Errorf("expected PushPop(42) = 42 on empty heap, got %d", got)
	}
	if h.Len() != 0 {
		t.Errorf("expected Len() = 0 after PushPop on empty, got %d", h.Len())
	}
}

func TestPushPop_EquivalentToPushThenPop(t *testing.T) {
	// Verify PushPop gives the same result as Push followed by Pop.
	rng := rand.New(rand.NewSource(12345))

	for trial := 0; trial < 100; trial++ {
		n := rng.Intn(20) + 1
		items := make([]int, n)
		for i := range items {
			items[i] = rng.Intn(1000)
		}
		pushVal := rng.Intn(1000)

		// Method 1: PushPop
		h1 := NewMinFrom(slices.Clone(items))
		got1 := h1.PushPop(pushVal)

		// Method 2: Push then Pop
		h2 := NewMinFrom(slices.Clone(items))
		h2.Push(pushVal)
		got2 := h2.Pop()

		if got1 != got2 {
			t.Errorf("trial %d: PushPop(%d) = %d, but Push+Pop = %d (items=%v)",
				trial, pushVal, got1, got2, items)
		}
	}
}

// ---------------------------------------------------------------------------
// Replace
// ---------------------------------------------------------------------------

func TestReplace(t *testing.T) {
	h := NewMin[int]()
	h.Push(1)
	h.Push(5)
	h.Push(3)

	old := h.Replace(10)
	if old != 1 {
		t.Errorf("expected Replace to return old root 1, got %d", old)
	}
	assertHeapInvariant(t, h)

	if got := h.Peek(); got != 3 {
		t.Errorf("expected new root 3, got %d", got)
	}
}

func TestReplace_PanicsOnEmpty(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on Replace on empty heap")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "Replace") {
			t.Errorf("unexpected panic value: %v", r)
		}
	}()

	h := NewMin[int]()
	h.Replace(1)
}

func TestReplace_EquivalentToPopThenPush(t *testing.T) {
	rng := rand.New(rand.NewSource(54321))

	for trial := 0; trial < 100; trial++ {
		n := rng.Intn(20) + 1
		items := make([]int, n)
		for i := range items {
			items[i] = rng.Intn(1000)
		}
		replaceVal := rng.Intn(1000)

		// Method 1: Replace
		h1 := NewMinFrom(slices.Clone(items))
		got1 := h1.Replace(replaceVal)

		// Method 2: Pop then Push
		h2 := NewMinFrom(slices.Clone(items))
		got2 := h2.Pop()
		h2.Push(replaceVal)

		if got1 != got2 {
			t.Errorf("trial %d: Replace(%d) = %d, but Pop+Push = %d",
				trial, replaceVal, got1, got2)
		}
	}
}

// ---------------------------------------------------------------------------
// Remove
// ---------------------------------------------------------------------------

func TestRemove_Found(t *testing.T) {
	h := NewMinFrom([]int{5, 3, 8, 1, 4, 7, 2})

	removed := h.Remove(func(v int) bool { return v == 4 })
	if !removed {
		t.Error("expected Remove to return true for existing element")
	}
	if h.Len() != 6 {
		t.Errorf("expected Len() = 6, got %d", h.Len())
	}
	assertHeapInvariant(t, h)

	// Verify 4 is no longer in the heap.
	if h.Contains(func(v int) bool { return v == 4 }) {
		t.Error("removed element 4 still found in heap")
	}
}

func TestRemove_NotFound(t *testing.T) {
	h := NewMinFrom([]int{1, 2, 3})

	removed := h.Remove(func(v int) bool { return v == 99 })
	if removed {
		t.Error("expected Remove to return false for non-existing element")
	}
	if h.Len() != 3 {
		t.Errorf("expected Len() = 3, got %d", h.Len())
	}
}

func TestRemove_Root(t *testing.T) {
	h := NewMinFrom([]int{1, 5, 3, 7, 9})

	removed := h.Remove(func(v int) bool { return v == 1 })
	if !removed {
		t.Error("expected Remove to return true for root")
	}
	assertHeapInvariant(t, h)
	if h.Peek() != 3 {
		t.Errorf("expected new root 3, got %d", h.Peek())
	}
}

func TestRemove_LastElement(t *testing.T) {
	h := NewMin[int]()
	h.Push(42)

	removed := h.Remove(func(v int) bool { return v == 42 })
	if !removed {
		t.Error("expected Remove to return true")
	}
	if !h.IsEmpty() {
		t.Error("expected empty heap after removing last element")
	}
}

func TestRemove_OnlyFirstOccurrence(t *testing.T) {
	h := NewMinFrom([]int{3, 3, 3})

	h.Remove(func(v int) bool { return v == 3 })
	if h.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d (Remove should only remove first match)", h.Len())
	}
}

func TestRemove_MaintainsInvariant_Randomized(t *testing.T) {
	rng := rand.New(rand.NewSource(99999))

	for trial := 0; trial < 50; trial++ {
		n := rng.Intn(50) + 5
		items := make([]int, n)
		for i := range items {
			items[i] = rng.Intn(100)
		}
		h := NewMinFrom(items)

		target := items[rng.Intn(n)]
		h.Remove(func(v int) bool { return v == target })
		assertHeapInvariant(t, h)
	}
}

// ---------------------------------------------------------------------------
// Contains
// ---------------------------------------------------------------------------

func TestContains(t *testing.T) {
	h := NewMinFrom([]int{1, 2, 3, 4, 5})

	if !h.Contains(func(v int) bool { return v == 3 }) {
		t.Error("expected Contains to return true for 3")
	}
	if h.Contains(func(v int) bool { return v == 99 }) {
		t.Error("expected Contains to return false for 99")
	}
}

func TestContains_EmptyHeap(t *testing.T) {
	h := NewMin[int]()
	if h.Contains(func(v int) bool { return v == 1 }) {
		t.Error("expected Contains to return false on empty heap")
	}
}

// ---------------------------------------------------------------------------
// Len / IsEmpty / Clear
// ---------------------------------------------------------------------------

func TestLen(t *testing.T) {
	h := NewMin[int]()

	if h.Len() != 0 {
		t.Errorf("expected Len() = 0, got %d", h.Len())
	}
	h.Push(1)
	if h.Len() != 1 {
		t.Errorf("expected Len() = 1, got %d", h.Len())
	}
	h.Push(2)
	if h.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d", h.Len())
	}
	h.Pop()
	if h.Len() != 1 {
		t.Errorf("expected Len() = 1 after Pop, got %d", h.Len())
	}
}

func TestIsEmpty(t *testing.T) {
	h := NewMin[int]()
	if !h.IsEmpty() {
		t.Error("expected IsEmpty() = true for new heap")
	}
	h.Push(1)
	if h.IsEmpty() {
		t.Error("expected IsEmpty() = false after Push")
	}
	h.Pop()
	if !h.IsEmpty() {
		t.Error("expected IsEmpty() = true after Pop")
	}
}

func TestClear(t *testing.T) {
	h := NewMinFrom([]int{1, 2, 3, 4, 5})
	h.Clear()

	if !h.IsEmpty() {
		t.Error("expected IsEmpty() = true after Clear")
	}
	if h.Len() != 0 {
		t.Errorf("expected Len() = 0 after Clear, got %d", h.Len())
	}

	// Heap should be usable again after Clear.
	h.Push(10)
	if h.Len() != 1 {
		t.Errorf("expected Len() = 1 after Push post-Clear, got %d", h.Len())
	}
	if h.Peek() != 10 {
		t.Errorf("expected Peek() = 10, got %d", h.Peek())
	}
}

// ---------------------------------------------------------------------------
// Values
// ---------------------------------------------------------------------------

func TestValues(t *testing.T) {
	items := []int{5, 3, 8, 1, 4}
	h := NewMinFrom(items)

	values := h.Values()
	if len(values) != h.Len() {
		t.Errorf("expected len(Values()) = %d, got %d", h.Len(), len(values))
	}

	// Values should contain all original items (regardless of order).
	slices.Sort(values)
	expected := slices.Clone(items)
	slices.Sort(expected)
	if !slices.Equal(values, expected) {
		t.Errorf("Values() = %v, want all of %v", values, expected)
	}
}

func TestValues_IndependentCopy(t *testing.T) {
	h := NewMinFrom([]int{1, 2, 3})
	values := h.Values()

	// Modifying the returned slice should not affect the heap.
	values[0] = 999
	if h.Peek() == 999 {
		t.Error("modifying Values() result should not affect the heap")
	}
}

func TestValues_EmptyHeap(t *testing.T) {
	h := NewMin[int]()
	values := h.Values()

	if len(values) != 0 {
		t.Errorf("expected empty slice, got %v", values)
	}
}

// ---------------------------------------------------------------------------
// Sorted
// ---------------------------------------------------------------------------

func TestSorted_MinHeap(t *testing.T) {
	h := NewMinFrom([]int{5, 3, 8, 1, 4, 7, 2, 9, 6})

	sorted := h.Sorted()
	if !slices.IsSorted(sorted) {
		t.Errorf("Sorted() result is not sorted: %v", sorted)
	}
	if len(sorted) != h.Len() {
		t.Error("Sorted() should not modify the original heap")
	}
}

func TestSorted_MaxHeap(t *testing.T) {
	h := NewMaxFrom([]int{5, 3, 8, 1, 4})

	sorted := h.Sorted()
	// For a max-heap, Sorted() should return descending order.
	for i := 1; i < len(sorted); i++ {
		if sorted[i] > sorted[i-1] {
			t.Errorf("Sorted() for max-heap not in descending order: %v", sorted)
			break
		}
	}
}

func TestSorted_DoesNotModifyHeap(t *testing.T) {
	h := NewMinFrom([]int{3, 1, 2})
	origLen := h.Len()

	_ = h.Sorted()

	if h.Len() != origLen {
		t.Errorf("Sorted() modified the heap: Len() changed from %d to %d", origLen, h.Len())
	}
}

func TestSorted_Empty(t *testing.T) {
	h := NewMin[int]()
	sorted := h.Sorted()
	if len(sorted) != 0 {
		t.Errorf("expected empty slice from Sorted on empty heap, got %v", sorted)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString(t *testing.T) {
	h := NewMinFrom([]int{3, 1, 2})
	s := h.String()

	if !strings.HasPrefix(s, "Heap[") {
		t.Errorf("expected String() to start with 'Heap[', got %q", s)
	}
}

func TestString_Empty(t *testing.T) {
	h := NewMin[int]()
	s := h.String()
	if s != "Heap[]" {
		t.Errorf("expected 'Heap[]' for empty heap, got %q", s)
	}
}

// ---------------------------------------------------------------------------
// Custom comparator / struct types
// ---------------------------------------------------------------------------

type task struct {
	name     string
	priority int
}

func TestNew_CustomStruct(t *testing.T) {
	h := New(func(a, b task) bool {
		return a.priority < b.priority
	})

	h.Push(task{"low", 10})
	h.Push(task{"high", 1})
	h.Push(task{"medium", 5})

	got := h.Pop()
	if got.name != "high" || got.priority != 1 {
		t.Errorf("expected highest-priority task, got %+v", got)
	}
	assertHeapInvariant(t, h)
}

func TestNew_StableRelativeOrder(t *testing.T) {
	// When priorities are equal, the heap should still function correctly.
	h := New(func(a, b task) bool {
		return a.priority < b.priority
	})

	h.Push(task{"a", 1})
	h.Push(task{"b", 1})
	h.Push(task{"c", 1})

	// All three should be extractable without error.
	for i := 0; i < 3; i++ {
		got := h.Pop()
		if got.priority != 1 {
			t.Errorf("expected priority 1, got %d", got.priority)
		}
	}
	if !h.IsEmpty() {
		t.Error("expected empty heap")
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestPush_LargeNumber(t *testing.T) {
	h := NewMin[int]()
	n := 10_000

	for i := n; i > 0; i-- {
		h.Push(i)
	}

	if h.Peek() != 1 {
		t.Errorf("expected Peek() = 1, got %d", h.Peek())
	}
	if h.Len() != n {
		t.Errorf("expected Len() = %d, got %d", n, h.Len())
	}
	assertHeapInvariant(t, h)
}

func TestPop_DrainCompletely(t *testing.T) {
	h := NewMinFrom([]int{5, 3, 8, 1, 4, 7, 2, 9, 6})

	for !h.IsEmpty() {
		h.Pop()
	}

	if h.Len() != 0 {
		t.Errorf("expected Len() = 0, got %d", h.Len())
	}
}

func TestPushPop_SingleElement(t *testing.T) {
	h := NewMin[int]()
	h.Push(5)

	// PushPop with a value equal to the root.
	got := h.PushPop(5)
	if got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
	if h.Len() != 1 {
		t.Errorf("expected Len() = 1, got %d", h.Len())
	}
}

func TestRemove_EmptyHeap(t *testing.T) {
	h := NewMin[int]()
	removed := h.Remove(func(v int) bool { return v == 1 })
	if removed {
		t.Error("expected Remove to return false on empty heap")
	}
}

// ---------------------------------------------------------------------------
// Randomized stress test
// ---------------------------------------------------------------------------

func TestRandomized_MinHeap(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	h := NewMin[int]()

	var reference []int

	for i := 0; i < 1000; i++ {
		switch rng.Intn(3) {
		case 0, 1: // Push (more likely)
			v := rng.Intn(10000)
			h.Push(v)
			reference = append(reference, v)
		case 2: // Pop
			if h.IsEmpty() {
				continue
			}
			got := h.Pop()
			slices.Sort(reference)
			expected := reference[0]
			reference = reference[1:]
			if got != expected {
				t.Fatalf("iteration %d: Pop() = %d, want %d", i, got, expected)
			}
		}
		assertHeapInvariant(t, h)
	}
}

func TestRandomized_MaxHeap(t *testing.T) {
	rng := rand.New(rand.NewSource(7777))
	h := NewMax[int]()

	var reference []int

	for i := 0; i < 1000; i++ {
		switch rng.Intn(3) {
		case 0, 1:
			v := rng.Intn(10000)
			h.Push(v)
			reference = append(reference, v)
		case 2:
			if h.IsEmpty() {
				continue
			}
			got := h.Pop()
			slices.Sort(reference)
			expected := reference[len(reference)-1]
			reference = reference[:len(reference)-1]
			if got != expected {
				t.Fatalf("iteration %d: Pop() = %d, want %d", i, got, expected)
			}
		}
		assertHeapInvariant(t, h)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkPush(b *testing.B) {
	h := NewMin[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
	}
}

func BenchmarkPop(b *testing.B) {
	h := NewMin[int]()
	for i := 0; i < b.N; i++ {
		h.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Pop()
	}
}

func BenchmarkPushPop(b *testing.B) {
	h := NewMinFrom(make([]int, 1000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.PushPop(i)
	}
}

func BenchmarkNewMinFrom(b *testing.B) {
	items := make([]int, 10000)
	rng := rand.New(rand.NewSource(0))
	for i := range items {
		items[i] = rng.Intn(100000)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMinFrom(items)
	}
}

// ---------------------------------------------------------------------------
// fmt.Stringer interface compliance
// ---------------------------------------------------------------------------

func TestStringer_Interface(t *testing.T) {
	h := NewMin[int]()
	h.Push(1)

	// Verify Heap satisfies fmt.Stringer via Sprintf %s.
	s := fmt.Sprintf("%s", h)
	if s == "" {
		t.Error("expected non-empty string from Stringer")
	}
}
