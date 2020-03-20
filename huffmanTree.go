package main

type (
	Node struct {
		weight uint64
		left   *Node
		right  *Node
		char   uint8
	}

	huffmanTree struct {
		minHeap []Node
		len     int
		codes   map[uint8]string
	}
)

func swap(a, b *Node) {
	*a, *b = *b, *a
}

func (hmTree *huffmanTree) minHeapify(i int) {
	if 2*i+1 >= hmTree.len {
		return
	}

	min := i

	if hmTree.minHeap[min].weight > hmTree.minHeap[2*i+1].weight {
		min = 2*i + 1
	}

	if 2*i+2 < hmTree.len && hmTree.minHeap[min].weight > hmTree.minHeap[2*i+2].weight {
		min = 2*i + 2
	}

	if min != i {
		swap(&hmTree.minHeap[min], &hmTree.minHeap[i])
		hmTree.minHeapify(min)
	}
}

func createNode(weight uint64, char uint8) Node {
	var nodeTemp Node
	nodeTemp.weight = weight
	nodeTemp.char = char

	return nodeTemp
}

func (hmTree *huffmanTree) addNode(node Node, flag bool) {
	hmTree.minHeap = append(hmTree.minHeap, node)

	hmTree.len++

	if flag {
		swap(&hmTree.minHeap[hmTree.len-1], &hmTree.minHeap[len(hmTree.minHeap)-1])

		if node.weight < hmTree.minHeap[(hmTree.len-1)/2].weight {
			for i := (hmTree.len - 1) / 2; i >= 0; i-- {
				hmTree.minHeapify(i)
			}
		}
	}
}

func (hmTree *huffmanTree) buildMinHeap(freq []uint64) {
	for i := 0; i < 256; i++ {
		if freq[i] > 0 {
			var node Node = createNode(freq[i], uint8(i))
			hmTree.addNode(node, false)
		}
	}

	for i := (hmTree.len - 1) / 2; i >= 0; i-- {
		hmTree.minHeapify(i)
	}

	for hmTree.len > 1 {
		a := hmTree.getMinNode()
		b := hmTree.getMinNode()

		var node Node
		node.weight = a.weight + b.weight
		node.char = '#'

		if a.weight <= b.weight {
			node.left = &a
			node.right = &b
		} else {
			node.left = &b
			node.right = &a
		}

		hmTree.addNode(node, true)
	}
}

func (hmTree *huffmanTree) getMinNode() Node {
	var nodeTemp Node = hmTree.minHeap[0]

	swap(&hmTree.minHeap[0], &hmTree.minHeap[hmTree.len-1])

	hmTree.len--
	hmTree.minHeapify(0)

	return nodeTemp
}

func (hmTree *huffmanTree) getCharCode(top *Node, code string) {
	if top == nil {
		return
	}

	if isLeaf(top) {
		hmTree.codes[top.char] = code
	}

	hmTree.getCharCode(top.left, code+"0")
	hmTree.getCharCode(top.right, code+"1")
}

func isLeaf(node *Node) bool {
	return (node.left == nil && node.right == nil)
}

func (hmTree *huffmanTree) getAllCode() {
	hmTree.codes = make(map[uint8]string)
	hmTree.getCharCode(&hmTree.minHeap[0], "")
}
