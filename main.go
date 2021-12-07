package main

import (
	"errors"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/cover"

	"github.com/nikolaydubina/treemap"
	"github.com/nikolaydubina/treemap/render"
)

const doc string = `
Generate heat treemaps for cover Go cover profile.

Example:

$ go test -coverprofile cover.out ./...
$ go-cover-heatmap -coverprofile cover.out > out.svg

Command options:
`

var grey = color.RGBA{128, 128, 128, 255}

func main() {
	var (
		coverprofile    string
		w               float64
		h               float64
		marginBox       float64
		paddingBox      float64
		padding         float64
		imputeHeat      bool
		countStatements bool
	)

	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), doc)
		flag.PrintDefaults()
	}
	flag.StringVar(&coverprofile, "coverprofile", "", "filename of input coverprofile (e.g. cover.out)")
	flag.Float64Var(&w, "w", 1028, "width of output")
	flag.Float64Var(&h, "h", 640, "height of output")
	flag.Float64Var(&marginBox, "margin-box", 4, "margin between boxes")
	flag.Float64Var(&paddingBox, "padding-box", 4, "padding between box border and content")
	flag.Float64Var(&padding, "padding", 32, "padding around root content")
	flag.BoolVar(&imputeHeat, "impute-heat", true, "impute heat for parents(weighted sum) and leafs(0.5)")
	flag.BoolVar(&countStatements, "statements", true, "count statemtents in files for size of files, when false then each file is size 1")
	flag.Parse()

	if coverprofile == "" {
		log.Fatal("coverprofile argument is missing")
	}

	profiles, err := cover.ParseProfiles(coverprofile)
	if err != nil {
		log.Fatal(err)
	}

	tree, err := coverageTreemap(profiles, countStatements)
	if err != nil {
		log.Fatal(err)
	}

	sizeImputer := treemap.SumSizeImputer{EmptyLeafSize: 1}
	sizeImputer.ImputeSize(*tree)

	if imputeHeat {
		heatImputer := treemap.WeightedHeatImputer{EmptyLeafHeat: 0.5}
		heatImputer.ImputeHeat(*tree)
	}

	// Note, we should not normalize heat since go coverage already reports 0~100%.

	palette, ok := render.GetPalette("RdYlGn")
	if !ok {
		log.Fatalf("can not get palette")
	}
	uiBuilder := render.UITreeMapBuilder{
		Colorer:     render.HeatColorer{Palette: palette},
		BorderColor: grey,
	}
	spec := uiBuilder.NewUITreeMap(*tree, w, h, marginBox, paddingBox, padding)
	renderer := render.SVGRenderer{}

	os.Stdout.Write(renderer.Render(spec, w, h))
}

// This is based on official go tool.
// Returns value in range 0~1
// Official reference: https://github.com/golang/go/blob/master/src/cmd/cover/html.go#L97
func percentCovered(p *cover.Profile) float64 {
	var total, covered int64
	for _, b := range p.Blocks {
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total)
}

func numStatements(p *cover.Profile) int {
	var total int
	for _, b := range p.Blocks {
		total += b.NumStmt
	}
	return total
}

// coverageTreemap create single treemap tree where each leaf is a file
// heat is test coverage
// size is number of lines
func coverageTreemap(profiles []*cover.Profile, countStatements bool) (*treemap.Tree, error) {
	if len(profiles) == 0 {
		return nil, errors.New("no profiles passed")
	}
	tree := treemap.Tree{
		Nodes: map[string]treemap.Node{},
		To:    map[string][]string{},
	}

	// for finding roots
	hasParent := map[string]bool{}

	for _, profile := range profiles {
		if profile == nil {
			return nil, fmt.Errorf("got nil profile")
		}

		if _, ok := tree.Nodes[profile.FileName]; ok {
			return nil, fmt.Errorf("duplicate node(%s)", profile.FileName)
		}

		var size int = 1
		if countStatements {
			size = numStatements(profile)
			if size == 0 {
				// fallback
				size = 1
			}
		}

		tree.Nodes[profile.FileName] = treemap.Node{
			Path:    profile.FileName,
			Size:    float64(size),
			Heat:    percentCovered(profile),
			HasHeat: true,
		}

		parts := strings.Split(profile.FileName, "/")
		hasParent[parts[0]] = false

		for parent, i := parts[0], 1; i < len(parts); i++ {
			child := parent + "/" + parts[i]

			tree.To[parent] = append(tree.To[parent], child)
			hasParent[child] = true

			parent = child
		}
	}

	for node, v := range tree.To {
		tree.To[node] = unique(v)
	}

	var roots []string
	for node, has := range hasParent {
		if !has {
			roots = append(roots, node)
		}
	}

	switch {
	case len(roots) == 0:
		return nil, errors.New("no roots, possible cycle in graph")
	case len(roots) > 1:
		tree.Root = "some-secret-string"
		tree.To[tree.Root] = roots
	default:
		tree.Root = roots[0]
	}

	return &tree, nil
}

func unique(a []string) []string {
	u := map[string]bool{}
	var b []string
	for _, q := range a {
		if _, ok := u[q]; !ok {
			u[q] = true
			b = append(b, q)
		}
	}
	return b
}
