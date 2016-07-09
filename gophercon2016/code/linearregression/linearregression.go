// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0

// This example demonstrates how to train a regression model in Go.  The example
// also prints out formatted results and saves two plot: (i) a plot of the raw input
// data, and (ii) a plot of the trained function overlaid on the raw input data.
// The input data is data about Go github repositories gathered via getrepos.go.
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pkg/errors"
	"github.com/sajari/regression"
)

func main() {

	// Aggregate the counts of created repos per day over all days,
	// in the input data set "repodata.csv."
	counts, err := prepareCountData("repodata.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Create and save the first plot showing the time series of
	// the raw input data.
	xys := preparePlotData(counts)
	if err = makePlots(xys); err != nil {
		log.Fatal(err)
	}

	// Perform a regression analysis and print the results for inspection.
	r := performRegression(counts)
	fmt.Printf("Regression:\n%s\n", r)

	// Generate the data for the second plot.
	xysPredicted, err := prepareRegPlotData(r, counts)
	if err != nil {
		log.Fatal(err)
	}

	// Create and save the second plot.
	if err = makeRegPlots(xys, xysPredicted); err != nil {
		log.Fatal(err)
	}

	// Make a prediction for the number of Go repositories that will
	// be created on the first day of GopherCon (July 11, 2016, or 1287 days
	// from the start of our data set).
	gcValue, err := r.Predict([]float64{1287.0})
	if err != nil {
		log.Fatal(err)
	}

	// Display the results.
	fmt.Printf("Day of GopherCon Prediction: %d\n", int(gcValue))
}

// prepareCountData prepares the raw time series data for plotting.
func prepareCountData(filename string) ([][]int, error) {

	// Store the daily counts of created repos.
	countMap := make(map[int]int)

	// Open the raw data file.
	csvBuffer, err := getFileFromPach("repodata.csv", "master", "godata")
	if err != nil {
		return [][]int{}, errors.Wrap(err, "Could not open CSV file")
	}

	// Parse the csv data in repodata.csv.
	reader := csv.NewReader(bytes.NewReader(csvBuffer.Bytes()))
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return [][]int{}, errors.Wrap(err, "Could not read in raw CSV data")
	}

	// Create a map of daily created repos where the keys are the days and
	// the values are the counts of created repos on that day.
	startTime := time.Date(2013, time.January, 1, 0, 0, 0, 0, time.UTC)
	layout := "2006-01-02 15:04:05"
	for _, each := range rawCSVdata {
		t, err := time.Parse(layout, each[2][0:19])
		if err != nil {
			return [][]int{}, errors.Wrap(err, "Could not parse timestamps")
		}
		interval := int(t.Sub(startTime).Hours() / 24.0)
		countMap[interval]++
	}

	// Sort the day values which is required for plotting.
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

// getFileFromPach gets the repodata.csv file pachyderm data versioning
func getFileFromPach(filename, branch, repoName string) (bytes.Buffer, error) {

	// Open a connection to pachyderm running on localhost.
	c, err := client.NewFromAddress("localhost:30650")
	if err != nil {
		return bytes.Buffer{}, errors.Wrap(err, "Could not connect to Pachyderm")
	}

	// Read the latest commit of filename to the given repoName.
	var buffer bytes.Buffer
	if err := c.GetFile(repoName, branch, filename, 0, 0, "", nil, &buffer); err != nil {
		return buffer, errors.Wrap(err, "Could not retrieve pachyderm file")
	}

	return buffer, nil
}

// preparePlotData prepares the raw input data for plotting.
func preparePlotData(counts [][]int) plotter.XYs {
	pts := make(plotter.XYs, len(counts))
	var i int

	for _, count := range counts {
		pts[i].X = float64(count[0])
		pts[i].Y = float64(count[1])
		i++
	}

	return pts
}

// preformRegression performs a linear regression of create repo counts vs. day.
func performRegression(counts [][]int) *regression.Regression {
	var r regression.Regression
	r.SetObserved("count of created Github repos")
	r.SetVar(0, "days since Jan 1 2013")

	for _, count := range counts {
		r.Train(regression.DataPoint(
			float64(count[1]),
			[]float64{float64(count[0])}))
	}

	r.Run()
	return &r
}

// prepareRegPlotData prepares predicted point for plotting.
func prepareRegPlotData(r *regression.Regression, counts [][]int) (plotter.XYs, error) {
	pts := make(plotter.XYs, len(counts))
	var i int

	for _, count := range counts {
		pts[i].X = float64(count[0])
		value, err := r.Predict([]float64{float64(count[0])})
		if err != nil {
			return pts, errors.Wrap(err, "Could not calculate predicted value")
		}
		pts[i].Y = value
		i++
	}

	return pts, nil
}

// makeRegPlots makes the second plot including the raw input data and the
// trained function.
func makeRegPlots(xys1, xys2 plotter.XYs) error {

	// Create a plot value.
	p, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not create plot object")
	}

	// Label the plot.
	p.Title.Text = "Daily Counts of Go Repos Created"
	p.X.Label.Text = "Days from Jan. 1, 2013"
	p.Y.Label.Text = "Count"

	// Add both sets of points, predicted and actual, to the plot.
	if err := plotutil.AddLinePoints(p, "Actual", xys1, "Predicted", xys2); err != nil {
		return errors.Wrap(err, "Could not add lines to plot")
	}

	// Save the plot.
	if err := p.Save(7*vg.Inch, 4*vg.Inch, "regression.png"); err != nil {
		return errors.Wrap(err, "Could not output plot")
	}

	return nil
}

// makePlots creates and saves the first of our plots showing the raw input data.
func makePlots(xys plotter.XYs) error {

	// Create a new plot.
	p, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not create plot object")
	}

	// Label the new plot.
	p.Title.Text = "Daily Counts of Go Repos Created"
	p.X.Label.Text = "Days from Jan. 1, 2013"
	p.Y.Label.Text = "Count"

	// Add the prepared points to the plot.
	if err = plotutil.AddLinePoints(p, "Counts", xys); err != nil {
		return errors.Wrap(err, "Could not add lines to plot")
	}

	// Save the plot to a PNG file.
	if err := p.Save(7*vg.Inch, 4*vg.Inch, "countseries.png"); err != nil {
		return errors.Wrap(err, "Could not output plot")
	}

	return nil
}
