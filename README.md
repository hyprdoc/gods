# gods

**Go Data Structures** — A type-safe, generic data structure library for Go.

`gods` provides production-ready implementations of fundamental data structures using Go generics. No `interface{}`, no type assertions, no code generation — just clean, strongly-typed APIs that work with any type.

```
go get github.com/hyprdoc/gods
```

> Requires **Go 1.21+** (generics with `cmp.Ordered` support).

---

## Why gods?

Go's standard library is deliberately minimal when it comes to data structures. The `container/heap` package, for example, requires implementing an interface with five methods and relies on `interface{}` — making it verbose and error-prone. `gods` takes a different approach:

- **Type-safe generics** — No casting. The compiler catches type errors at build time.
- **Zero-alloc-aware** — Popped/dequeued slots are zeroed to prevent GC leaks from stale references in backing arrays.
- **Consistent API** — Every structure follows the same patterns: `Try*` variants, `Values()`, `Clear()`, `String()`, `IsEmpty()`.
- **No dependencies** — Only the Go standard library.

---

## Data Structures

| Package | Structure | Backing | Description |
|---|---|---|---|
| [`heap`](#heap) | Binary Heap | Slice | Min-heap, max-heap, or custom ordering |
| [`stack`](#stack) | LIFO Stack | Slice | Last-in, first-out |
| [`queue`](#queue) | FIFO Queue | Ring buffer | First-in, first-out with auto grow/shrink |

---

## Heap

A binary heap with configurable ordering. Supports min-heap, max-heap, and arbitrary comparator functions for custom types.

```go
import "github.com/hyprdoc/gods/heap"
```

### Quick start

```go
// Min-heap of integers
h := heap.NewMin[int]()
h.Push(5)
h.Push(2)
h.Push(8)

fmt.Println(h.Pop())  // 2
fmt.Println(h.Peek()) // 5
```

### Max-heap

```go
h := heap.NewMax[int]()
h.Push(5)
h.Push(2)
h.Push(8)

fmt.Println(h.Pop()) // 8
```

### Custom comparator

Use `heap.New` for types that aren't `cmp.Ordered`, or when you need custom ordering:

```go
type Task struct {
    Name     string
    Priority int
}

h := heap.New(func(a, b Task) bool {
    return a.Priority < b.Priority // lower number = higher priority
})

h.Push(Task{"Send email", 3})
h.Push(Task{"Fix bug", 1})
h.Push(Task{"Update docs", 2})

fmt.Println(h.Pop().Name) // "Fix bug"
```

### Build from a slice (O(n))

When you already have a collection of elements, building the heap from a slice is faster than pushing one-by-one. `NewMinFrom` / `NewMaxFrom` use Floyd's heap construction algorithm — **O(n)** instead of O(n log n).

```go
scores := []int{88, 42, 97, 63, 55}
h := heap.NewMinFrom(scores) // O(n) construction

fmt.Println(h.Pop()) // 42
```

### Efficient combined operations

`PushPop` and `Replace` avoid redundant sift operations, making them ideal for streaming top-K problems:

```go
// Keep the 3 largest scores seen so far
top3 := heap.NewMin[int]()

for _, score := range allScores {
    if top3.Len() < 3 {
        top3.Push(score)
    } else {
        top3.PushPop(score) // push + pop in one O(log n) pass
    }
}
```

### API

| Method | Time | Description |
|---|---|---|
| `New(less)` | — | Custom comparator heap |
| `NewMin()` / `NewMax()` | — | Ordered-type min/max heap |
| `NewFrom(items, less)` | O(n) | Build from slice |
| `NewMinFrom(items)` / `NewMaxFrom(items)` | O(n) | Build from ordered-type slice |
| `Push(value)` | O(log n) | Add element |
| `Pop()` | O(log n) | Remove and return root (panics if empty) |
| `TryPop()` | O(log n) | Safe pop — returns `(T, bool)` |
| `Peek()` | O(1) | View root without removing (panics if empty) |
| `TryPeek()` | O(1) | Safe peek — returns `(T, bool)` |
| `PushPop(value)` | O(log n) | Push then pop in one operation |
| `Replace(value)` | O(log n) | Pop then push in one operation |
| `Remove(equal)` | O(n) | Remove first match by predicate |
| `Contains(predicate)` | O(n) | Search by predicate |
| `Sorted()` | O(n log n) | Sorted copy (heap unchanged) |
| `Len()` / `IsEmpty()` | O(1) | Size queries |
| `Values()` | O(n) | Unordered copy |
| `Clear()` | O(1) | Release all elements |

---

## Stack

A classic LIFO (last-in, first-out) stack backed by a Go slice.

```go
import "github.com/hyprdoc/gods/stack"
```

### Quick start

```go
s := stack.New[string]()
s.Push("a")
s.Push("b")
s.Push("c")

fmt.Println(s.Pop())  // "c"
fmt.Println(s.Peek()) // "b"
fmt.Println(s.Len())  // 2
```

### Build from a slice

The last element of the input slice becomes the top of the stack:

```go
s := stack.NewFrom([]int{1, 2, 3})
fmt.Println(s.Pop()) // 3
```

### Pre-allocate capacity

When you know the approximate size upfront, avoid repeated slice growth:

```go
s := stack.NewWithCapacity[int](1000)
```

### API

| Method | Time | Description |
|---|---|---|
| `New()` | — | Empty stack |
| `NewWithCapacity(n)` | — | Pre-allocated stack |
| `NewFrom(items)` | O(n) | Build from slice (last = top) |
| `Push(value)` | O(1)* | Add to top |
| `Pop()` | O(1) | Remove and return top (panics if empty) |
| `TryPop()` | O(1) | Safe pop — returns `(T, bool)` |
| `Peek()` | O(1) | View top without removing (panics if empty) |
| `TryPeek()` | O(1) | Safe peek — returns `(T, bool)` |
| `Len()` / `IsEmpty()` | O(1) | Size queries |
| `Values()` | O(n) | Copy from bottom to top |
| `Clear()` | O(1) | Release all elements |

*amortized

---

## Queue

A FIFO (first-in, first-out) queue backed by a **growable ring buffer**.

```go
import "github.com/hyprdoc/gods/queue"
```

### Quick start

```go
q := queue.New[int]()
q.Enqueue(1)
q.Enqueue(2)
q.Enqueue(3)

fmt.Println(q.Dequeue()) // 1
fmt.Println(q.Peek())    // 2
fmt.Println(q.PeekBack()) // 3
```

### Build from a slice

Elements are enqueued in slice order — `items[0]` is at the front:

```go
q := queue.NewFrom([]string{"first", "second", "third"})
fmt.Println(q.Dequeue()) // "first"
```

### Why a ring buffer?

A naive slice-based queue using `data = data[1:]` to dequeue has a subtle memory leak: the underlying array still holds references to dequeued elements, preventing garbage collection. This is especially dangerous with pointer types or large structs.

The ring buffer in `gods/queue` solves this by:
1. **Reusing slots** — head and tail indices wrap around the buffer.
2. **Zeroing dequeued positions** — explicitly clearing references for GC.
3. **Auto-shrinking** — halving the buffer when occupancy drops below 25%, so memory isn't held unnecessarily after a burst.

### API

| Method | Time | Description |
|---|---|---|
| `New()` | — | Empty queue |
| `NewWithCapacity(n)` | — | Pre-allocated (rounds to power of 2) |
| `NewFrom(items)` | O(n) | Build from slice (items[0] = front) |
| `Enqueue(value)` | O(1)* | Add to back |
| `Dequeue()` | O(1)* | Remove and return front (panics if empty) |
| `TryDequeue()` | O(1)* | Safe dequeue — returns `(T, bool)` |
| `Peek()` | O(1) | View front without removing (panics if empty) |
| `TryPeek()` | O(1) | Safe peek — returns `(T, bool)` |
| `PeekBack()` | O(1) | View back element (panics if empty) |
| `TryPeekBack()` | O(1) | Safe peek back — returns `(T, bool)` |
| `Len()` / `IsEmpty()` | O(1) | Size queries |
| `Values()` | O(n) | Copy from front to back |
| `Clear()` | O(n) | Zero all slots and release |

*amortized

---

## Design Decisions

### Panicking vs. returning errors

`Pop`, `Peek`, and `Dequeue` panic on empty containers rather than returning `(T, error)`. This matches the convention of Go's built-in operations (e.g., indexing a nil map key is fine, but indexing out-of-bounds panics). Popping an empty container is always a programming error — not a recoverable runtime condition.

For callers who prefer checked access, every panicking method has a `Try*` variant that returns `(T, bool)`.

### GC-safe zeroing

When an element is popped, dequeued, or removed, its slot in the backing array is explicitly zeroed:

```go
var zero T
s.data[top] = zero
```

Without this, the backing array retains a reference to the removed object. For pointer types, slices, maps, or structs containing them, this prevents garbage collection — a classic Go memory leak that's easy to miss.

### Ring buffer for Queue

A slice-based queue that dequeues with `data = data[1:]` appears to work, but:
- The backing array never reclaims front slots, growing memory monotonically.
- Even with periodic re-slicing, dequeued slots hold stale references.

The power-of-two ring buffer avoids both issues, enables fast modular arithmetic via bitwise AND, and automatically shrinks when underutilized.

### Consistent API surface

All three structures share a common interface pattern:

- `Len()` / `IsEmpty()` — size queries
- `Clear()` — reset to empty
- `Values()` — returns an independent copy
- `String()` — debug-friendly output (implements `fmt.Stringer`)
- `Try*` — safe alternatives to panicking methods

This consistency makes the library predictable and easy to learn.

---

## Thread Safety

None of these data structures are safe for concurrent use. This is a deliberate choice — embedding a mutex adds overhead for all users, even single-threaded ones. If you need concurrent access, wrap calls with a `sync.Mutex`:

```go
var mu sync.Mutex
h := heap.NewMin[int]()

mu.Lock()
h.Push(42)
mu.Unlock()
```

---

## Roadmap

This library is under active development. Planned additions include:

- **Data structures** — Linked list, deque, ordered map, ordered set, trie, graph
- **Algorithms** — Sorting, searching, graph traversal
- **Benchmarks** — Comparative benchmarks against standard library alternatives

Contributions and suggestions are welcome.

---

## License

[MIT](LICENSE) © hyprdoc
