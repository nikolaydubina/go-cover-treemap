package covertreemap

import (
	"strconv"
	"strings"

	"github.com/nikolaydubina/treemap"
)

func AddCoveragePercentageToName(tree *treemap.Tree) {
	for path, node := range tree.Nodes {
		var builder strings.Builder

		builder.WriteString(node.Name)
		builder.WriteString(" ")
		builder.WriteString(strconv.FormatFloat(node.Heat*100, 'f', 1, 64))
		builder.WriteString("%")

		tree.Nodes[path] = treemap.Node{
			Path:    node.Path,
			Name:    builder.String(),
			Size:    node.Size,
			Heat:    node.Heat,
			HasHeat: node.HasHeat,
		}
	}
}
