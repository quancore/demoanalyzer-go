package analyser

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"

	"github.com/pforemski/gouda/point"
	"github.com/quancore/demoanalyzer-go/utils"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

// SimplePlotter is the default standard plotter for 2-dimensional data sets
type SimplePlotter struct {
}

// A monokai-ish color palette
var colors = []drawing.Color{
	drawing.ColorFromHex("f92672"),
	drawing.ColorFromHex("89bdff"),
	drawing.ColorFromHex("66d9ef"),
	drawing.ColorFromHex("67210c"),
	drawing.ColorFromHex("7acd10"),
	drawing.ColorFromHex("af619f"),
	drawing.ColorFromHex("fd971f"),
	drawing.ColorFromHex("dcc060"),
	drawing.ColorFromHex("545250"),
	drawing.ColorFromHex("4b7509"),
}

// PointsInDimension returns all coordinates in a given dimension
func (p SimplePlotter) PointsInDimension(points point.Points, n int) []float64 {
	var v []float64
	for _, p := range points {
		if len(p.V) > n {
			v = append(v, p.V[n])
		}

	}
	return v
}

// ClusterCenter return cluster center in given cluster
func (p SimplePlotter) ClusterCenter(points point.Points) *point.Point {
	mean := points.Mean()
	return mean
}

// Plot draw a 2-dimensional data set into a PNG file named {k_iteration}.png
func (p SimplePlotter) Plot(outPath, mapname string, bounds image.Rectangle, cc []point.Points) {
	var series []chart.Series
	var clusterCenters point.Points

	// draw data points
	for i, c := range cc {
		if len(c) > 0 {
			series = append(series, chart.ContinuousSeries{
				Style: chart.Style{
					Show:        true,
					StrokeWidth: chart.Disabled,
					DotColor:    colors[i%len(colors)],
					DotWidth:    8,
				},
				XValues: p.PointsInDimension(c, 0),
				YValues: p.PointsInDimension(c, 1),
			})

			clusterCenter := p.ClusterCenter(c)
			clusterCenters = append(clusterCenters, clusterCenter)
		}

	}

	// draw cluster center points
	// series = append(series, chart.ContinuousSeries{
	// 	Style: chart.Style{
	// 		Show:        true,
	// 		StrokeWidth: chart.Disabled,
	// 		DotColor:    drawing.ColorBlack,
	// 		DotWidth:    16,
	// 	},
	// 	XValues: p.PointsInDimension(clusterCenters, 0),
	// 	YValues: p.PointsInDimension(clusterCenters, 1),
	// })

	graph := chart.Chart{
		Height: bounds.Dy(),
		Width:  bounds.Dx(),
		Background: chart.Style{
			FillColor: drawing.Color{R: 0, G: 0, B: 0, A: 1},
		},
		Canvas: chart.Style{
			FillColor: drawing.Color{R: 0, G: 0, B: 0, A: 1},
		},
		Series: series,
	}

	// graph := chart.Chart{
	// 	Height: 1024,
	// 	Width:  1024,
	// 	Series: series,
	// }

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	utils.CheckError(err)

	clusterImg, _, _ := image.Decode(bytes.NewReader(buffer.Bytes()))

	ioutil.WriteFile(fmt.Sprintf("cluster.png"), buffer.Bytes(), 0644)

	// read base radar map
	gopath := utils.GetGoPath()
	fMap, err := os.Open(fmt.Sprintf("%s/src/github.com/markus-wa/demoinfocs-golang/metadata/maps/%s.jpg", gopath, mapname))
	utils.CheckError(err)
	imgMap, _, err := image.Decode(fMap)
	utils.CheckError(err)

	// Create output canvas and use map overview image as base
	img := image.NewRGBA(imgMap.Bounds())
	draw.Draw(img, imgMap.Bounds(), imgMap, image.ZP, draw.Over)

	// Generate and draw heatmap overlay on top of the overview
	draw.Draw(img, bounds, clusterImg, image.ZP, draw.Over)

	// Write to stdout
	f, o_err := os.Create(outPath)
	utils.CheckError(o_err)

	defer f.Close()
	err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpegQuality})
	utils.CheckError(err)

}
