// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0
//
// This example demonstrates how to create histograms with gonum/plot.  The program
// saves two histograms.  The first is a histogram of github star values per Go repo
// (see getrepos.go for information on how this data is retrieved), and the second
// is a histogram of the log of non-zero star values.
//
package main

import (
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/pkg/errors"
)

func main() {

	// First, we prepare our input data for plotting.
	v, vl, err := prepareStarData("repodata.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Next, we create and save the histogram plots.
	if err = makePlots(v, vl); err != nil {
		log.Fatal(err)
	}
}

// makePlots creates and saves both histogram plots.
func makePlots(v, vl plotter.Values) error {

	// Plots are created here, and then we set its title.
	p1, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not generate a new plot")
	}
	p1.Title.Text = "Histogram of Github Stars"
	p2, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "Could not generate a new plot")
	}
	p2.Title.Text = "Histogram of log(Github Stars)"

	// The histograms are then created and added to the respective plots p1 and p2.
	h1, err := plotter.NewHist(v, 16)
	if err != nil {
		return errors.Wrap(err, "Could not create histogram")
	}
	p1.Add(h1)
	h2, err := plotter.NewHist(vl, 16)
	if err != nil {
		return errors.Wrap(err, "Could not create histogram")
	}
	p2.Add(h2)

	// The plots are then saved in the current directory.
	if err := p1.Save(4*vg.Inch, 4*vg.Inch, "hist1.png"); err != nil {
		return errors.Wrap(err, "Could not save plot")
	}
	if err := p2.Save(4*vg.Inch, 4*vg.Inch, "hist2.png"); err != nil {
		return errors.Wrap(err, "Could not save plot")
	}

	return nil

}

// prepareStartData translates the input CSV data into values for gonum/plot
func prepareStarData(filename string) (plotter.Values, plotter.Values, error) {

	// Our raw CSV file in opened.
	csvfile, err := os.Open("repodata.csv")
	if err != nil {
		return plotter.Values{}, plotter.Values{}, errors.Wrap(err, "Could not open CSV file")
	}
	defer csvfile.Close()

	// The csv data is extracted from the open file.
	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return plotter.Values{}, plotter.Values{}, errors.Wrap(err, "Could not read in raw CSV data")
	}

	// We then loop over each row of the CSV data adding the data into the plotting "Values."
	v := make(plotter.Values, len(rawCSVdata))
	var vl plotter.Values
	for i, each := range rawCSVdata {
		value, err := strconv.ParseFloat(each[5], 64)
		if err != nil {
			return plotter.Values{}, plotter.Values{}, errors.Wrap(err, "Could not convert value to float")
		}
		v[i] = value
		if value != 0.0 {
			vl = append(vl, math.Log(value))
		}
	}

	return v, vl, nil

}
