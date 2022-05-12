package covertreemap

import (
	"strings"

	"github.com/nikolaydubina/treemap"
)

// RemoveFilesTreeFilter removes files from Tree.
// Sizes and parents have to be already imputed and have sizes and heat.
func RemoveFilesTreeFilter(tree *treemap.Tree, substr string) {
	if substr == "" {
		return
	}

	for key := range tree.Nodes {
		if strings.HasSuffix(key, substr) {
			delete(tree.Nodes, key)
		}
	}

	for parent, children := range tree.To {
		childrenNew := make([]string, 0, len(children))

		for _, child := range children {
			if !strings.HasSuffix(child, substr) {
				childrenNew = append(childrenNew, child)
			}
		}

		tree.To[parent] = childrenNew
	}
}
