package search

type PriorityItem interface {
	Priority() float64
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []PriorityItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority() > pq[j].Priority()
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	// n := len(*pq)
	item := x.(PriorityItem)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
// func (pq *PriorityQueue) update(item PriorityItem, value string, priority int) {
// 	heap.Remove(pq, item.index)
// 	item.value = value
// 	item.priority = priority
// 	heap.Push(pq, item)
// }
