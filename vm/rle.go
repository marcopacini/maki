package vm

import "fmt"

type node struct {
	Value int
	Count int
	next  *node
}

type RLE struct {
	head *node
	tail *node
	size int
}

func NewRLE() *RLE {
	return &RLE{
		head: nil,
		tail: nil,
	}
}

func (r *RLE) Add(val int) {
	if r.tail != nil {
		if r.tail.Value == val {
			r.tail.Count++
		} else {
			n := &node{
				Value: val,
				Count: 1,
				next:  nil,
			}

			r.tail.next = n
			r.tail = n
		}
	} else {
		n := &node{
			Value: val,
			Count: 1,
			next:  nil,
		}

		r.head = n
		r.tail = n
	}
}

func (r RLE) At(i int) (int, error) {
	if i < 0 {
		return 0, fmt.Errorf("out of range")
	}

	n := r.head

	for n != nil {
		i -= n.Count
		if i < 0 {
			return n.Value, nil
		}
		n = n.next
	}

	return 0, fmt.Errorf("out of range")
}
