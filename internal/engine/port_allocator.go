package engine

import "sync"

type PortAllocator struct {
	mu        sync.Mutex
	minimum   int
	maximum   int
	next      int
	allocated map[int]struct{}
}

func NewPortAllocator(minimum int, maximum int) *PortAllocator {

	if minimum < 1 {
		minimum = 1
	}

	if maximum < minimum {
		maximum = minimum
	}

	return &PortAllocator{minimum: minimum, maximum: maximum, next: minimum, allocated: make(map[int]struct{})}
}

func (a *PortAllocator) Allocate(requested int) (int, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if requested > 0 {

		if requested < a.minimum || requested > a.maximum {
			return 0, false
		}

		if _, exists := a.allocated[requested]; exists {
			return 0, false
		}

		a.allocated[requested] = struct{}{}
		return requested, true
	}

	span := a.maximum - a.minimum + 1

	for offset := 0; offset < span; offset++ {
		port := a.minimum + (a.next-a.minimum+offset)%span

		if _, exists := a.allocated[port]; !exists {
			a.allocated[port] = struct{}{}
			a.next = port + 1

			if a.next > a.maximum {
				a.next = a.minimum
			}

			return port, true
		}
	}

	return 0, false
}

func (a *PortAllocator) Release(port int) {
	a.mu.Lock()
	delete(a.allocated, port)
	a.mu.Unlock()
}

func (a *PortAllocator) InUse(port int) bool {
	a.mu.Lock()
	_, exists := a.allocated[port]
	a.mu.Unlock()
	return exists
}
