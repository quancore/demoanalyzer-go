package analyser

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"os"

	heatmap "github.com/dustin/go-heatmap"
	"github.com/dustin/go-heatmap/schemes"
	"github.com/golang/geo/r2"
	cluster "github.com/pforemski/gouda/dbscan"
	point "github.com/pforemski/gouda/point"
	utils "github.com/quancore/demoanalyzer-go/utils"
)

const (
	dotSize     = 15
	opacity     = 128
	jpegQuality = 90
	epsX        = 40
	epsY        = 40
	minpoints   = 4
	outImg      = "heatmap.jpg"
	clusterOut  = "cluster.png"
)

func (analyser *Analyser) printHeadmap() {
	var positions []r2.Point
	for _, position := range analyser.killPositions {
		positions = append(positions, position.KillPoint)
	}
	r2Bounds := r2.RectFromPoints(positions...)
	bounds := image.Rectangle{
		Min: image.Point{X: int(r2Bounds.X.Lo), Y: int(r2Bounds.Y.Lo)},
		Max: image.Point{X: int(r2Bounds.X.Hi), Y: int(r2Bounds.Y.Hi)},
	}

	// Transform r2.Points into heatmap.DataPoints
	var data []heatmap.DataPoint
	for _, p := range positions[1:] {
		// Invert Y since go-heatmap expects data to be ordered from bottom to top
		data = append(data, heatmap.P(p.X, p.Y*-1))
	}

	// read base radar map
	gopath := utils.GetGoPath()
	fMap, err := os.Open(fmt.Sprintf("%s/src/github.com/markus-wa/demoinfocs-golang/metadata/maps/%s.jpg", gopath, analyser.mapName))
	utils.CheckError(err)
	imgMap, _, err := image.Decode(fMap)
	utils.CheckError(err)

	// Create output canvas and use map overview image as base
	img := image.NewRGBA(imgMap.Bounds())
	draw.Draw(img, imgMap.Bounds(), imgMap, image.ZP, draw.Over)

	// Generate and draw heatmap overlay on top of the overview
	imgHeatmap := heatmap.Heatmap(image.Rect(0, 0, bounds.Dx(), bounds.Dy()), data, dotSize, opacity, schemes.AlphaFire)
	draw.Draw(img, bounds, imgHeatmap, image.ZP, draw.Over)

	// Write to stdout
	f, err := os.Create(outImg)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpegQuality})
	utils.CheckError(err)

}

func (analyser *Analyser) clusterPoints() {
	var positions []r2.Point
	// var clusterPositions cluster.PointList
	var clusterPositions point.Points

	for _, position := range analyser.killPositions {
		positions = append(positions, position.KillPoint)
		newPoint := point.New(position.KillPoint.X, position.KillPoint.Y*-1)
		// clusterPositions = append(clusterPositions, cluster.Point{position.KillPoint.X, position.KillPoint.Y})
		clusterPositions = append(clusterPositions, newPoint)

	}
	r2Bounds := r2.RectFromPoints(positions...)
	bounds := image.Rectangle{
		Min: image.Point{X: int(r2Bounds.X.Lo), Y: int(r2Bounds.Y.Lo)},
		Max: image.Point{X: int(r2Bounds.X.Hi), Y: int(r2Bounds.Y.Hi)},
	}

	// clusters, noise := cluster.DBScan(clusterPositions, 0.08, 10) // eps is 800m, 10 points minimum in eps-neighborhood
	clusters := cluster.Search(clusterPositions, []float64{epsX, epsY}, minpoints)
	plotter := SimplePlotter{}
	plotter.Plot(clusterOut, analyser.mapName, bounds, clusters)

}
