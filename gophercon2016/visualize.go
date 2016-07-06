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

	// prepare our data
	v, vl, err := prepareStarData("repodata.csv")
	if err != nil {
		log.Fatal(err)
	}

	// save plots
	err = makePlots(v, vl)
	if err != nil {
		log.Fatal(err)
	}
}

func makePlots(v, vl plotter.Values) error {

	// Make a plot and set its title.
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

	// Create a histogram of our values
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

	// Save the plot to a PNG file.
	if err := p1.Save(4*vg.Inch, 4*vg.Inch, "hist1.png"); err != nil {
		return errors.Wrap(err, "Could not save plot")
	}
	if err := p2.Save(4*vg.Inch, 4*vg.Inch, "hist2.png"); err != nil {
		return errors.Wrap(err, "Could not save plot")
	}

	return nil

}

func prepareStarData(filename string) (plotter.Values, plotter.Values, error) {

	// Get our csv data
	csvfile, err := os.Open("repodata.csv")
	if err != nil {
		return plotter.Values{}, plotter.Values{}, errors.Wrap(err, "Could not open CSV file")
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return plotter.Values{}, plotter.Values{}, errors.Wrap(err, "Could not read in raw CSV data")
	}

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
		//vl[i] = math.Log(value)
	}

	return v, vl, nil

}
