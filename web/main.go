package main

import (
	_ "embed"
	"encoding/base64"
	"image/color"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"syscall/js"

	"golang.org/x/tools/cover"

	"github.com/nikolaydubina/go-cover-treemap/covertreemap"
	"github.com/nikolaydubina/treemap"
	"github.com/nikolaydubina/treemap/render"
)

var grey = color.RGBA{128, 128, 128, 255}

type Renderer struct {
	w          float64 // SVG width
	h          float64 // SVG height
	marginBox  float64
	paddingBox float64
	padding    float64
	fileText   string
	scale      int // this is how many % we multiply width and height of SVG
	hEpsilon   int // used to avoid scroll bar
}

func (r *Renderer) OnWindowResize(_ js.Value, _ []js.Value) interface{} {
	windowWidth := js.Global().Get("innerWidth").Int()
	windowHeight := js.Global().Get("innerHeight").Int()

	document := js.Global().Get("document")
	outputContainer := document.Call("getElementById", "output-container")
	fileInput := document.Call("getElementById", "file-input")

	w := windowWidth
	h := windowHeight - (outputContainer.Get("offsetTop").Int() - fileInput.Get("offsetHeight").Int()) - r.hEpsilon

	var f float64 = 1
	if r.scale > 0 {
		f = float64(r.scale) / 100
	}
	r.w = float64(w) * f
	r.h = float64(h) * f

	r.Render()
	return false
}

func (r *Renderer) OnDetailsSliderInputChange(_ js.Value, _ []js.Value) interface{} {
	document := js.Global().Get("document")
	s := document.Call("getElementById", "details-slider-input").Get("value").String()
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	r.scale = v
	r.OnWindowResize(js.Value{}, nil)
	r.Render()
	return false
}

func (r *Renderer) OnFileDrop(_ js.Value, args []js.Value) interface{} {
	event := args[0]
	event.Call("preventDefault")

	fileReader := js.Global().Get("FileReader").New()
	fileReader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		r.fileText = e.Get("target").Get("result").String()
		r.Render()
		return nil
	}))

	file := event.Get("dataTransfer").Get("files").Index(0)
	fileReader.Call("readAsText", file)

	return false
}

func (r *Renderer) OnDragOver(_ js.Value, _ []js.Value) interface{} {
	document := js.Global().Get("document")
	document.Call("getElementById", "file-input").Set("className", "file-input-hover")
	return false
}

func (r *Renderer) OnDragEnd(_ js.Value, _ []js.Value) interface{} {
	document := js.Global().Get("document")
	document.Call("getElementById", "file-input").Set("className", "")
	r.OnWindowResize(js.Value{}, nil)
	return false
}

func (r *Renderer) NewOnClickExample(examplePath string) func(this js.Value, args []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		go func() {
			resp, err := http.Get(examplePath)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			r.fileText = string(b)

			r.OnDragEnd(js.Value{}, nil)
			r.Render()
		}()
		return false
	}
}

func (r *Renderer) Render() {
	if r.fileText == "" {
		return
	}

	profiles, err := cover.ParseProfilesFromReader(strings.NewReader(r.fileText))
	if err != nil {
		log.Fatal(err)
	}

	treemapBuilder := covertreemap.NewCoverageTreemapBuilder(true)
	tree, err := treemapBuilder.CoverageTreemapFromProfiles(profiles)
	if err != nil {
		log.Fatal(err)
	}

	sizeImputer := treemap.SumSizeImputer{EmptyLeafSize: 1}
	sizeImputer.ImputeSize(*tree)
	treemap.SetNamesFromPaths(tree)
	treemap.CollapseLongPaths(tree)

	heatImputer := treemap.WeightedHeatImputer{EmptyLeafHeat: 0.5}
	heatImputer.ImputeHeat(*tree)

	palette, ok := render.GetPalette("RdYlGn")
	if !ok {
		log.Fatalf("can not get palette")
	}
	uiBuilder := render.UITreeMapBuilder{
		Colorer:     render.HeatColorer{Palette: palette},
		BorderColor: grey,
	}
	spec := uiBuilder.NewUITreeMap(*tree, r.w, r.h, r.marginBox, r.paddingBox, r.padding)
	renderer := render.SVGRenderer{}

	img := renderer.Render(spec, r.w, r.h)

	document := js.Global().Get("document")
	document.Call("getElementById", "output-container").Set("innerHTML", string(img))
	document.Call("getElementById", "file-input").Get("style").Set("display", "none")
	document.Call("getElementById", "details-slider-input-container").Get("style").Set("display", "")

	downloadButton := document.Call("getElementById", "download-button")
	downloadButton.Set("href", "data:image/svg;base64,"+base64.StdEncoding.EncodeToString(img))
	downloadButton.Set("download", "coverprofile-treemap.svg")
}

func main() {
	c := make(chan bool)
	renderer := Renderer{
		marginBox:  4,
		paddingBox: 4,
		padding:    16,
		hEpsilon:   16,
	}

	document := js.Global().Get("document")
	fileInput := document.Call("getElementById", "file-input")

	fileInput.Set("ondragover", js.FuncOf(renderer.OnDragOver))
	fileInput.Set("ondragend", js.FuncOf(renderer.OnDragEnd))
	fileInput.Set("ondragleave", js.FuncOf(renderer.OnDragEnd))
	fileInput.Set("ondrop", js.FuncOf(renderer.OnFileDrop))

	document.Call("getElementById", "example-chi").Set("onclick", js.FuncOf(renderer.NewOnClickExample("/go-cover-treemap/testdata/chi.cover")))
	document.Call("getElementById", "example-gin").Set("onclick", js.FuncOf(renderer.NewOnClickExample("/go-cover-treemap/testdata/gin.cover")))
	document.Call("getElementById", "example-hugo").Set("onclick", js.FuncOf(renderer.NewOnClickExample("/go-cover-treemap/testdata/hugo.cover")))

	document.Call("getElementById", "details-slider-input").Set("oninput", js.FuncOf(renderer.OnDetailsSliderInputChange))

	js.Global().Set("onresize", js.FuncOf(renderer.OnWindowResize))

	renderer.OnWindowResize(js.Value{}, nil)

	<-c
}
