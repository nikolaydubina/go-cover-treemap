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

		var builder strings.Builder

		if path == tree.Root {
			builder.WriteString(node.Name)
		} else {
			builder.WriteString(parts[len(parts)-1])
		}

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
