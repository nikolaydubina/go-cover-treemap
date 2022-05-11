package covertreemap

import (
	"math"
	"strings"

	"github.com/nikolaydubina/treemap"
)

type Filter struct {
	DirectoryOnly bool
}

// FilterTree filters the given tree applying the  current filters
func (f Filter) FilterTree(tree *treemap.Tree) {
	// Check if directory only filter is set
	if f.DirectoryOnly {
		f.filterDirectoriesOnly(tree)
	}
}

// filterDirectoriesOnly removes the node leafs and keep only the
// parent directories
func (f Filter) filterDirectoriesOnly(tree *treemap.Tree) {
	newNodes := make(map[string]treemap.Node, 0)
	newTo := make(map[string][]string, 0)

	for path, node := range tree.Nodes {
		if !strings.Contains(path, ".go") {
			var heat float64 = 0
			var size float64 = 0
			var nodesCount float64 = float64(len(tree.To[path]))
			if nodesCount == 0 {
				nodesCount = 1
			}

			for _, nodeId := range tree.To[path] {
				heat += tree.Nodes[nodeId].Heat
				size += tree.Nodes[nodeId].Size
			}

			newNodes[path] = treemap.Node{
				Path:    node.Path,
				Name:    node.Name,
				Size:    math.Log(size / nodesCount),
				Heat:    heat / nodesCount,
				HasHeat: heat > 0,
			}
		}
	}

	for path, nodePaths := range tree.To {
		newNodePaths := make([]string, 0)

		for _, nodePath := range nodePaths {
			if !strings.Contains(nodePath, ".go") {
				newNodePaths = append(newNodePaths, nodePath)
			}
		}

		newTo[path] = newNodePaths
	}

	tree.Nodes = newNodes
	tree.To = newTo
}
