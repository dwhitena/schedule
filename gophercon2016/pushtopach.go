package main

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pkg/errors"
)

func main() {

	err := pushToPach("repodata.csv")
	if err != nil {
		log.Fatal(err)
	}

}

func pushToPach(filename string) error {

	csvfile, err := ioutil.ReadFile(filename)
	if err != nil {
		errors.Wrap(err, "Could not read input file")
	}

	// Connect to pachyderm
	c, err := client.NewFromAddress("localhost:30650")
	if err != nil {
		errors.Wrap(err, "Could not connect to Pachyderm")
	}

	// Create a repo called "filer"
	if err := c.CreateRepo("filter"); err != nil {
		errors.Wrap(err, "Could not create pachyderm repo")
	}

	// Start a commit in our new repo on the "master" branch
	_, err = c.StartCommit("filter", "", "master")
	if err != nil {
		errors.Wrap(err, "Could not start pachyderm repo commit")
	}

	r := bytes.NewReader(csvfile)
	// Put a file in the newly created commit.
	if _, err := c.PutFile("filter", "master", filename, r); err != nil {
		errors.Wrap(err, "Could not put file into pachyderm repo")
	}

	// Finish the commit.
	if err := c.FinishCommit("filter", "master"); err != nil {
		errors.Wrap(err, "Could not finish Pachyderm commit")
	}

	return nil

}
