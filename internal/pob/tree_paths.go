package pob

import "unsafe"

// Поиск повторяет BuildPathFromNode из Classes/PassiveSpec.lua PoB v0.23.1.
// Copyright (c) 2016 David Gowor.
// Лицензия MIT сохранена в internal/pob/bridge/LICENSE.
const (
	pathAllocated uint32 = 1 << iota
	pathUnlocked
	pathMastery
	pathClassStart
	pathAllFlags = pathAllocated | pathUnlocked | pathMastery | pathClassStart
)

const maxPathWords = 1 << 24

type packedPathNode struct {
	ascendancy, mode, distance, flags uint32
	neighbors                         []uint32
}

type packedPathState struct {
	distance, root, tail uint32
}

type packedPaths struct {
	states    []packedPathState
	linkWords []uint32
	used      int
}

// Lua владеет обоими массивами. Go использует их только до возврата вызова.
func solvePathWords(in *uint32, inputSize uint32, out *uint32, outputSize uint32) uint32 {
	if in == nil || out == nil || inputSize < 3 || outputSize == 0 || inputSize > maxPathWords || outputSize > maxPathWords {
		return 0
	}
	input := unsafe.Slice(in, int(inputSize))
	nodes, roots, ok := unpackPathGraph(input)
	if !ok {
		return 0
	}
	return uint32(findPackedPaths(nodes, roots, unsafe.Slice(out, int(outputSize))))
}

func unpackPathGraph(input []uint32) ([]packedPathNode, []uint32, bool) {
	if len(input) < 3 {
		return nil, nil, false
	}
	nodeCount, edgeCount, rootCount := int(input[0]), int(input[1]), int(input[2])
	if nodeCount <= 0 || nodeCount > (len(input)-3)/5 || edgeCount < 0 || rootCount < 0 {
		return nil, nil, false
	}
	endOfNodes := 3 + 5*nodeCount
	indices := input[endOfNodes:]
	if edgeCount > len(indices) || len(indices)-edgeCount != rootCount || !pathIndicesFit(indices, uint32(nodeCount)) {
		return nil, nil, false
	}
	nodes, ok := unpackPathNodes(input[3:endOfNodes], indices[:edgeCount])
	return nodes, indices[edgeCount:], ok
}

func pathIndicesFit(indices []uint32, nodeCount uint32) bool {
	for _, index := range indices {
		if index >= nodeCount {
			return false
		}
	}
	return true
}

func unpackPathNodes(records, neighbors []uint32) ([]packedPathNode, bool) {
	nodes := make([]packedPathNode, len(records)/5)
	for index := range nodes {
		row := records[5*index:][:5]
		degree := int(row[4])
		if !validPackedPathNode(row, len(neighbors)) {
			return nil, false
		}
		nodes[index] = packedPathNode{
			ascendancy: row[0], mode: row[1], distance: row[2], flags: row[3], neighbors: neighbors[:degree],
		}
		neighbors = neighbors[degree:]
	}
	return nodes, len(neighbors) == 0
}

func validPackedPathNode(row []uint32, neighbors int) bool {
	return row[1] <= 2 && (row[2] == 0 || row[2] == 1000) &&
		row[3]&^pathAllFlags == 0 && row[4] <= uint32(neighbors)
}

func findPackedPaths(nodes []packedPathNode, roots, output []uint32) int {
	stateWords := 1 + 3*len(nodes)
	if len(output) < stateWords {
		return 0
	}
	paths := packedPaths{states: make([]packedPathState, len(nodes)), linkWords: output[stateWords:]}
	for index, node := range nodes {
		paths.states[index].distance = node.distance
	}
	queue := make([]uint32, 0, len(nodes))
	for _, root := range roots {
		var ok bool
		queue, ok = findPackedPathsFrom(nodes, root, queue[:0], &paths)
		if !ok {
			return 0
		}
	}
	output[0] = uint32(paths.used / 2)
	for index, state := range paths.states {
		position := 1 + 3*index
		output[position], output[position+1], output[position+2] = state.distance, state.root, state.tail
	}
	return stateWords + paths.used
}

func findPackedPathsFrom(nodes []packedPathNode, root uint32, queue []uint32, paths *packedPaths) ([]uint32, bool) {
	paths.states[root] = packedPathState{root: root + 1}
	rootMode := nodes[root].mode
	queue = append(queue, root)
	for head := 0; head < len(queue); head++ {
		var ok bool
		queue, ok = visitPackedNeighbors(nodes, root, rootMode, queue[head], queue, paths)
		if !ok {
			return nil, false
		}
	}
	return queue, true
}

func visitPackedNeighbors(nodes []packedPathNode, root, rootMode, index uint32, queue []uint32, paths *packedPaths) ([]uint32, bool) {
	node := nodes[index]
	if node.flags&(pathMastery|pathUnlocked) != pathUnlocked {
		return queue, true
	}
	state := paths.states[index]
	for _, next := range node.neighbors {
		other := nodes[next]
		if paths.states[next].distance <= state.distance || !canTraversePath(rootMode, node, other, state.distance) {
			continue
		}
		if !paths.add(next, root, state, other.flags&pathAllocated != 0) {
			return nil, false
		}
		queue = append(queue, next)
	}
	return queue, true
}

func canTraversePath(rootMode uint32, from, to packedPathNode, distance uint32) bool {
	if to.flags&(pathClassStart|pathUnlocked) != pathUnlocked {
		return false
	}
	if to.flags&pathAllocated != 0 && to.mode != 0 && to.mode != rootMode {
		return false
	}
	return from.ascendancy == to.ascendancy || (distance == 0 && to.ascendancy == 0)
}

func (p *packedPaths) add(index, root uint32, previous packedPathState, allocated bool) bool {
	if p.used+2 > len(p.linkWords) {
		return false
	}
	distance := previous.distance
	if !allocated {
		distance++
	}
	// Ссылка хранит прежний хвост, даже если при следующем корне он сменится.
	p.linkWords[p.used], p.linkWords[p.used+1] = index+1, previous.tail
	p.used += 2
	p.states[index] = packedPathState{distance: distance, root: root + 1, tail: uint32(p.used / 2)}
	return true
}
