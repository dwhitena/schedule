// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0
//
// This example demonstrates how to train a regression model in Go.  The example
// also prints out formatted results and saves two plot: (i) a plot of the raw input
// data, and (ii) a plot of the trained function overlaid on the raw input data.
// The input data is data about Go github repositories gathered via getrepos.go.
//
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

	// This step aggregates counts of created repos per day over all days,
	// in the input data set "repodata.csv."
	counts, err := prepareCountData("repodata.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Next, we create and save the first plot showing the time series of
	// the raw input data.
	xys := preparePlotData(counts)
	if err = makePlots(xys); err != nil {
		log.Fatal(err)
	}

	// A regression analysis is performed and the regression results are
	// printed for inspection.
	r := performRegression(counts)
	fmt.Printf("Regression formula:\n%v\n", r.Formula)
	fmt.Printf("Regression:\n%s\n", r)

	// We pass the regression object to another function that generated
	// the data for the second plot, and then we create and save the
	// second plot
	xysPredicted := prepareRegPlotData(r, counts)
	if err = makeRegPlots(xys, xysPredicted); err != nil {
		log.Fatal(err)
	}

	// Finally we make prediction for the number of Go repositories that will
	// be created on the first day of GopherCon (July 11, 2016, or 1287 days
	// from the start of our data set).
	gcValue, _ := r.Predict([]float64{1287.0})
	fmt.Printf("Day of GopherCon Prediction: %.2f\n", gcValue)

}

// makePlots creates and saves the first of our plots showing the raw input data.
func makePlots(xys plotter.XYs) error {

	// A new plot object is created.
	p, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not create plot object")
	}

	// Then we label the new plot.
	p.Title.Text = "Daily Counts of Go Repos Created"
	p.X.Label.Text = "Days from Jan. 1, 2013"
	p.Y.Label.Text = "Count"

	// Prepared point are then added to the plot.
	if err = plotutil.AddLinePoints(p, "Counts", xys); err != nil {
		return errors.Wrap(err, "Could not add lines to plot")
	}

	// Lastly, the plot is saved to a PNG file.
	if err := p.Save(7*vg.Inch, 4*vg.Inch, "countseries.png"); err != nil {
		return errors.Wrap(err, "Could not output plot")
	}

	return nil
}

// preparePlotData prepares the raw input data for plotting.
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

// makeRegPlots makes the second plot including the raw input data and the
// trained function.
func makeRegPlots(xys1, xys2 plotter.XYs) error {

	// A plot object is created.
	p, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not create plot object")
	}

	// Then we label the plot.
	p.Title.Text = "Daily Counts of Go Repos Created"
	p.X.Label.Text = "Days from Jan. 1, 2013"
	p.Y.Label.Text = "Count"

	// This time, we add both sets of points, predicted and actual, to the plot.
	err = plotutil.AddLinePoints(p,
		"Actual", xys1,
		"Predicted", xys2)
	if err != nil {
		return errors.Wrap(err, "Could not add lines to plot")
	}

	// Finally, the plot in saved.
	if err := p.Save(7*vg.Inch, 4*vg.Inch, "regression.png"); err != nil {
		return errors.Wrap(err, "Could not output plot")
	}

	return nil
}

// prepareRegPlotData prepares predicted point for plotting.
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

// prepareCountData prepares the raw time series data for plotting.
func prepareCountData(filename string) ([][]int, error) {

	// countMap will store daily counts of created repos.
	countMap := make(map[int]int)

	// We open the raw data.
	csvfile, err := os.Open("repodata.csv")
	if err != nil {
		return [][]int{}, errors.Wrap(err, "Could not open CSV file")
	}
	defer csvfile.Close()

	// Then we parse the csv data in repodata.csv.
	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return [][]int{}, errors.Wrap(err, "Could not read in raw CSV data")
	}

	// A map of daily created repos is created, where the keys are the days and
	// the values are the counts of created repos on that day.
	startTime := time.Date(2013, time.January, 1, 0, 0, 0, 0, time.UTC)
	layout := "2006-01-02 15:04:05"
	for _, each := range rawCSVdata {
		t, _ := time.Parse(layout, each[2][0:19])
		interval := int(t.Sub(startTime).Hours() / 24.0)
		countMap[interval]++
	}

	// Lastly, the day values are sorted, which is required for plotting.
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

// preformRegression performs a linear regression of create repo counts vs. day.
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
