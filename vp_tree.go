package search

import (
	"container/heap"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"
)

// VPTreeDistancer interface compares two items to be sorted in the VP-Tree
type VPTreeDistancer interface {
	// Distance returns the distance between two items satisfying the triangle
	// inequality
	Distance(a, b VPTreeItem) float64
}

// VPTreeItem interface provides a generic interface to support indexing
// different types
type VPTreeItem interface {
	SetNode(*VPTreeNode)
	GetNode() *VPTreeNode
	ShouldSkip(VPTreeItem) bool
	ApplyAffinity(float64, VPTreeItem) float64
}

type vpTreeComparator struct {
	Distancer VPTreeDistancer
}

func (v *vpTreeComparator) Less(a, b *VPTreeNode) bool {
	return a.dist < b.dist
}

type VPTreeNode struct {
	index                 int
	threshold, m, M, dist float64
	left, right           *VPTreeNode
	children              []int
	isLeaf                bool
	_dead                 bool
}

// IsDead returns if the node has been marked for deletion
func (v *VPTreeNode) IsDead() bool {
	return v._dead
}

type vpHeapItem struct {
	index        int
	dist         float64
	node, parent *VPTreeNode
}

func (v *vpHeapItem) Priority() float64 {
	return v.dist
}

func (v *vpHeapItem) Less(other *vpHeapItem) bool {
	return v.dist < other.dist
}

// VPTree is an instance of a vp-tree index
type VPTree struct {
	// Distancer will be invoked to calculate the distance between items
	Distancer  VPTreeDistancer
	comparator vpTreeComparator
	root       *VPTreeNode
	items      []VPTreeItem
	_deadIdx   []int
	mutex      sync.Mutex
	// MaxChildren int
}

// SetItems will (re)build the index for the slice of items
func (v *VPTree) SetItems(items []VPTreeItem) {
	v.items = items
	v._deadIdx = make([]int, 0)
	nodes := make([]*VPTreeNode, len(items))
	for i := 0; i < len(nodes); i++ {
		var n VPTreeNode
		item := items[i]
		n.index = i
		nodes[i] = &n
		item.SetNode(&n)
	}
	v.root = v.buildFromPoints(nodes)
}

// ItemCount returns the number of items in the tree
func (v *VPTree) ItemCount() int {
	return len(v.items)
}

// Search returns the nearest k items to the target. The items are sorted with
// by distance ascending. The second parameter is the repective distances to the
// target
func (v *VPTree) Search(target VPTreeItem, k int) ([]VPTreeItem, []float64) {

	tau := new(float64)
	*tau = math.MaxFloat64
	pq := &PriorityQueue{}
	heap.Init(pq)

	v.search(v.root, target, k, pq, tau, math.MaxFloat64, true)

	results := make([]VPTreeItem, pq.Len())
	distances := make([]float64, pq.Len())

	for i := pq.Len() - 1; i >= 0; i-- {
		item := heap.Pop(pq).(*vpHeapItem)
		results[i] = v.items[item.index]
		distances[i] = item.Priority()
	}

	return results, distances

}

// SearchInRange returns the nearest k items to the target sorted by distance
// ascending with no result being more that maxDistance away from the target.
func (v *VPTree) SearchInRange(target VPTreeItem, k int, maxDist float64) ([]VPTreeItem, []float64) {

	tau := new(float64)
	*tau = maxDist
	pq := &PriorityQueue{}
	heap.Init(pq)

	v.search(v.root, target, k, pq, tau, maxDist, true)

	results := make([]VPTreeItem, pq.Len())
	distances := make([]float64, pq.Len())

	for i := pq.Len() - 1; i >= 0; i-- {
		item := heap.Pop(pq).(*vpHeapItem)
		results[i] = v.items[item.index]
		distances[i] = item.Priority()
	}

	return results, distances

}

func (v *VPTree) search(node *VPTreeNode, target VPTreeItem, k int, pq *PriorityQueue, tau *float64, maxDist float64, applyAffinity bool) {
	if node == nil {
		return
	}

	// if node.isLeaf {
	// 	for _, idx := range node.children {
	// 		dist := v.Distancer.Distance(v.items[idx], target)
	// 		if dist < v.tau {
	// 			if pq.Len() == k {
	// 				heap.Pop(pq)
	// 			}
	// 			heap.Push(pq, &vpHeapItem{idx, dist, nil, node})
	// 			if pq.Len() == k {
	// 				v.tau = (*pq)[0].Priority()
	// 			}
	// 		}
	// 	}
	// 	return
	// }

	if node._dead || v.items[node.index].ShouldSkip(target) {
		v.search(node.left, target, k, pq, tau, maxDist, applyAffinity)
		v.search(node.right, target, k, pq, tau, maxDist, applyAffinity)
		return
	}

	dist := v.Distancer.Distance((v.items)[node.index], target)
	var priority float64
	if applyAffinity && dist < maxDist {
		priority = (v.items)[node.index].ApplyAffinity(dist, target)
	} else {
		priority = dist
	}
	t := *tau

	// This Vantage-point is close enough
	if priority < t {
		if pq.Len() == k {
			heap.Pop(pq)
		}

		heap.Push(pq, &vpHeapItem{
			index:  node.index,
			dist:   priority,
			node:   node,
			parent: nil})

		if pq.Len() == k {
			item := heap.Pop(pq).(*vpHeapItem)
			*tau = item.Priority()
			heap.Push(pq, item)
		}
	}

	if node.left == nil && node.right == nil {
		return
	}

	if dist < node.threshold {
		if node.left != nil && node.m-t <= dist {
			v.search(node.left, target, k, pq, tau, maxDist, applyAffinity)
		}
		if node.right != nil && node.threshold-t < dist && dist < node.M+t {
			v.search(node.right, target, k, pq, tau, maxDist, applyAffinity)
		}
	} else {
		if node.right != nil && node.m-t < dist {
			v.search(node.right, target, k, pq, tau, maxDist, applyAffinity)
		}
		if node.left != nil && node.m-t < dist && dist < node.threshold+t {
			v.search(node.left, target, k, pq, tau, maxDist, applyAffinity)
		}
	}
}

func (v *VPTree) medianOf3(list []*VPTreeNode, a int, b int, c int) int {
	A, B, C := list[a], list[b], list[c]
	if v.comparator.Less(A, B) {
		if v.comparator.Less(B, C) {
			return b
		}
		if v.comparator.Less(A, C) {
			return c
		}
		return a
	}
	if v.comparator.Less(A, C) {
		return a
	}
	if v.comparator.Less(B, C) {
		return c
	}
	return b
}

func (v *VPTree) partition(list []*VPTreeNode, left, right, pivotIndex int) int {
	pivotValue := list[pivotIndex]
	list[pivotIndex], list[right] = list[right], list[pivotIndex]
	storeIndex := left
	for i := left; i < right; i++ {
		if v.comparator.Less(list[i], pivotValue) {
			list[storeIndex], list[i] = list[i], list[storeIndex]
			storeIndex++
		}
	}
	list[right], list[storeIndex] = list[storeIndex], list[right]
	return storeIndex
}

func (v *VPTree) nthElement(list []*VPTreeNode, left, nth, right int) *VPTreeNode {
	var pivotIndex, pivotNewIndex, pivotDist int
	for {
		pivotIndex = v.medianOf3(list, left, right, (left+right)>>1)
		pivotNewIndex = v.partition(list, left, right, pivotIndex)
		pivotDist = pivotNewIndex - left + 1
		if pivotDist == nth {
			return list[pivotNewIndex]
		} else if nth < pivotDist {
			right = pivotNewIndex - 1
		} else {
			nth -= pivotDist
			left = pivotNewIndex + 1
		}
	}
}

func (v *VPTree) buildFromPoints(nodes []*VPTreeNode) *VPTreeNode {
	listLength := len(nodes)
	if listLength == 0 {
		return nil
	}

	// // Is this a leaf node
	// if tree.MaxChildren > 1 && delta <= tree.MaxChildren {
	// 	node.children = make([]int, delta)
	// 	node.isLeaf = true
	// 	for i := 0; i < delta-1; i++ {
	// 		node.children[i] = lower + i + 1
	// 	}
	// 	return &node
	// }
	vpIndex := rand.Intn(listLength)
	node := nodes[vpIndex]
	nodes = append(nodes[0:vpIndex], nodes[vpIndex+1:]...)
	listLength--
	if listLength == 0 {
		return node
	}

	vp := v.items[node.index]

	// Ensure Distance calculations are only done once per sort
	S := v.items
	var wg sync.WaitGroup
	distances := make([]float64, listLength)
	batchSize := int(math.Ceil(float64(listLength) / float64(runtime.NumCPU())))
	for i := 0; i < listLength; i += batchSize {
		wg.Add(batchSize)
		go func(idx int) {
			var batch []*VPTreeNode
			if idx+batchSize < listLength {
				batch = nodes[idx:(idx + batchSize)]
			} else {
				batch = nodes[idx:listLength]
			}
			for j, item := range batch {
				dist := v.Distancer.Distance(vp, S[item.index])
				item.dist = dist
				distances[j] = dist
			}
			wg.Add(-batchSize)
		}(i)
	}

	wg.Wait()
	sort.Float64s(distances)

	node.m = distances[0]
	node.M = distances[listLength-1]

	medianIndex := listLength >> 1
	median := v.nthElement(nodes, 0, medianIndex+1, listLength-1)

	leftItems := nodes[0:medianIndex]
	rightItems := nodes[medianIndex:]

	node.threshold = median.dist
	node.left = v.buildFromPoints(leftItems)
	node.right = v.buildFromPoints(rightItems)

	return node
}

// Insert adds a new item to the index
func (v *VPTree) Insert(item VPTreeItem) {

	if (len(v.items) - len(v._deadIdx)) <= 0 {
		v.SetItems([]VPTreeItem{item})
		return
	}

	tau := new(float64)
	*tau = math.MaxFloat64
	pq := &PriorityQueue{}
	heap.Init(pq)

	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.search(v.root, item, 1, pq, tau, math.MaxFloat64, false)

	heapItem := (*pq)[0].(*vpHeapItem)

	var match *VPTreeNode

	if heapItem.node != nil {
		match = heapItem.node
	} else {
		match = heapItem.parent
	}

	var node VPTreeNode
	node.index = len(v.items)
	items := append(v.items, item)
	v.items = items
	node.isLeaf = false
	item.SetNode(&node)

	for {
		dist := v.Distancer.Distance(v.items[match.index], item)
		if dist <= match.threshold {
			if dist < match.m {
				match.m = dist
			}
			if match.left == nil {
				match.m = dist
				match.left = &node
				return
			}
			match = match.left
		} else {
			if dist > match.M {
				match.M = dist
			}
			if match.right == nil {
				match.M = dist
				match.right = &node
				return
			}
			match = match.right
		}
	}
}

// Remove marks that an item should no longer be included in search results. The
// item will be removed from the index when the index rebuilds
func (v *VPTree) Remove(item VPTreeItem) {
	if v.root == nil {
		return
	}

	if node := item.GetNode(); node != nil {
		node._dead = true
		v.mutex.Lock()
		v._deadIdx = append(v._deadIdx, node.index)
		v.mutex.Unlock()
		return
	}

	tau := new(float64)
	*tau = math.MaxFloat64
	pq := &PriorityQueue{}
	heap.Init(pq)

	v.search(v.root, item, 1, pq, tau, math.MaxFloat64, false)

	if pq.Len() >= 1 {
		heapItem := (*pq)[0].(*vpHeapItem)

		var match *VPTreeNode

		if heapItem.node != nil {
			match = heapItem.node
		} else {
			match = heapItem.parent
		}

		match._dead = true
		v.mutex.Lock()
		v._deadIdx = append(v._deadIdx, match.index)
		v.mutex.Unlock()
	}
}

// Rebuild will trigger a rebuild on the index over the same items. All items
// marked for removal will be removed from the item list at this stage
func (v *VPTree) Rebuild() {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	sort.Ints(v._deadIdx)
	l := v.items
	for i := len(v._deadIdx) - 1; i >= 0; i-- {
		didx := v._deadIdx[i]
		l = append(l[0:didx], l[didx+1:]...)
	}
	v.SetItems(l)
}

func (v *VPTree) Items() []VPTreeItem {
	return v.items
}
