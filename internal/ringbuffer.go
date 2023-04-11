package internal

type RingBuffer struct {
	data    []string
	head    int
	tail    int
	size    int
	maxSize int
}

func NewRingBuffer(maxSize int) *RingBuffer {
	return &RingBuffer{
		data:    make([]string, maxSize),
		head:    0,
		tail:    0,
		size:    0,
		maxSize: maxSize,
	}
}

func (rb *RingBuffer) Put(item string) {
	rb.data[rb.head] = item
	rb.head = (rb.head + 1) % rb.maxSize
	if rb.size < rb.maxSize {
		rb.size++
	} else {
		rb.tail = (rb.tail + 1) % rb.maxSize
	}
}

func (rb *RingBuffer) Contains(item string) bool {
	for i := 0; i < rb.size; i++ {
		index := (rb.tail + i) % rb.maxSize
		if rb.data[index] == item {
			return true
		}
	}
	return false
}
