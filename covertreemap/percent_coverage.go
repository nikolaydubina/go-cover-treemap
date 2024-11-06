package covertreemap

import (
	"strconv"
	"strings"

	"github.com/nikolaydubina/treemap"
)

func AddCoveragePercentageToName(tree *treemap.Tree) {
	for path, node := range tree.Nodes {
		parts := strings.Split(node.Path, "/")
		if len(parts) == 0 {
			continue
		}

		tree.Nodes[path] = treemap.Node{
			Path:    node.Path,
			Name:    parts[len(parts)-1] + " " + strconv.FormatFloat(node.Heat*100, 'f', 1, 64) + "%",
			Size:    node.Size,
			Heat:    node.Heat,
			HasHeat: node.HasHeat,
		}
	}
}
