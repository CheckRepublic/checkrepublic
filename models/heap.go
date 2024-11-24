package models

// Define a max-heap
type MaxHeap []*Offer
func (h MaxHeap) Len() int           { return len(h) }
func (h MaxHeap) Less(i, j int) bool { 
	if h[i].Price == h[j].Price {
		return h[i].ID.String() < h[j].ID.String()
	}
	
	return h[i].Price > h[j].Price } // Inverted to make it a max-heap
func (h MaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(*Offer))
}

func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}


// Define a min-heap
type MinHeap []*Offer

func (h MinHeap) Len() int { return len(h) }

// Modify Less for min-heap behavior
func (h MinHeap) Less(i, j int) bool {
	if h[i].Price == h[j].Price {
		return h[i].ID.String() < h[j].ID.String()
	}

	return h[i].Price < h[j].Price // Regular comparison for min-heap
}

func (h MinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*Offer))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}