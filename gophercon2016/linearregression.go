package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/pkg/errors"
	"github.com/sajari/regression"
)

func main() {

	// prepare our data
	counts, err := prepareCountData("repodata.csv")
	if err != nil {
		log.Fatal(err)
	}

	// create time series plot
	xys := preparePlotData(counts)
	err = makePlots(xys)
	if err != nil {
		log.Fatal(err)
	}

	// perform regression
	r := performRegression(counts)
	fmt.Printf("Regression formula:\n%v\n", r.Formula)
	fmt.Printf("Regression:\n%s\n", r)

	// make regression plot
	xysPredicted := prepareRegPlotData(r, counts)
	err = makeRegPlots(xys, xysPredicted)
	if err != nil {
		log.Fatal(err)
	}

	// make gopherCon prediction
	gcValue, _ := r.Predict([]float64{1287.0})
	fmt.Printf("Day of GopherCon Prediction: %.2f\n", gcValue)

}

func makePlots(xys plotter.XYs) error {

	p, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not create plot object")
	}

	p.Title.Text = "Daily Counts of Go Repos Created"
	p.X.Label.Text = "Days from Jan. 1, 2013"
	p.Y.Label.Text = "Count"

	err = plotutil.AddLinePoints(p, "Counts", xys)
	if err != nil {
		return errors.Wrap(err, "Could not add lines to plot")
	}

	// Save the plot to a PNG file.
	if err := p.Save(7*vg.Inch, 4*vg.Inch, "countseries.png"); err != nil {
		return errors.Wrap(err, "Could not output plot")
	}

	return nil
}

func preparePlotData(counts [][]int) plotter.XYs {
	pts := make(plotter.XYs, len(counts))
	i := 0
	for _, count := range counts {
		pts[i].X = float64(count[0])
		pts[i].Y = float64(count[1])
		i++
	}
	return pts
}

func makeRegPlots(xys1, xys2 plotter.XYs) error {

	p, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not create plot object")
	}

	p.Title.Text = "Daily Counts of Go Repos Created"
	p.X.Label.Text = "Days from Jan. 1, 2013"
	p.Y.Label.Text = "Count"

	err = plotutil.AddLinePoints(p,
		"Actual", xys1,
		"Predicted", xys2)
	if err != nil {
		return errors.Wrap(err, "Could not add lines to plot")
	}

	// Save the plot to a PNG file.
	if err := p.Save(7*vg.Inch, 4*vg.Inch, "regression.png"); err != nil {
		return errors.Wrap(err, "Could not output plot")
	}

	return nil
}

func prepareRegPlotData(r *regression.Regression, counts [][]int) plotter.XYs {
	pts := make(plotter.XYs, len(counts))
	i := 0
	for _, count := range counts {
		pts[i].X = float64(count[0])
		value, _ := r.Predict([]float64{float64(count[0])})
		pts[i].Y = value
		i++
	}
	return pts
}

func prepareCountData(filename string) ([][]int, error) {

	countMap := make(map[int]int)

	// Get our csv data
	csvfile, err := os.Open("repodata.csv")
	if err != nil {
		return [][]int{}, errors.Wrap(err, "Could not open CSV file")
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return [][]int{}, errors.Wrap(err, "Could not read in raw CSV data")
	}

	// create a map of created repo counts per day
	startTime := time.Date(2013, time.January, 1, 0, 0, 0, 0, time.UTC)
	layout := "2006-01-02 15:04:05"
	for _, each := range rawCSVdata {
		t, _ := time.Parse(layout, each[2][0:19])
		interval := int(t.Sub(startTime).Hours() / 24.0)
		countMap[interval]++
	}

	// sort the values
	var keys []int
	for k := range countMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var sortedCounts [][]int
	for _, k := range keys {
		sortedCounts = append(sortedCounts, []int{k, countMap[k]})
	}

	return sortedCounts, nil

}

func performRegression(counts [][]int) *regression.Regression {
	r := new(regression.Regression)
	r.SetObserved("count of created Github repos")
	r.SetVar(0, "days since Jan 1 2013")
	for _, count := range counts {
		r.Train(
			regression.DataPoint(float64(count[1]), []float64{float64(count[0])}),
		)
	}
	r.Run()
	return r
}
