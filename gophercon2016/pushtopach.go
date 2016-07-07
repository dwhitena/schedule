// All material is licensed under the Apache License Version 2.0, January 2004
// http://www.apache.org/licenses/LICENSE-2.0
//
// This example demonstrates how to version a file in pachyderm's PFS file system.
// Note, the below logic assumes pachyderm is running on localhost.  However,
// localhost:30650 could be change to any <host>:30650 on which pachyderm is
// running.  See http://pachyderm.io/ for more details on getting started with
// pachyderm.
//
package main

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pkg/errors"
)

func main() {

	// Here we are pushing the file "repodata.csv" into pachyderm's PFS file system.
	if err := pushToPach("repodata.csv"); err != nil {
		log.Fatal(err)
	}

}

// pushToPach reads in the given file, create a repo in PFS for the file, opens
// a commit, pushes the file in the commit, and finishes the commit.
func pushToPach(filename string) error {

	// First we read the contents of the given file.
	csvfile, err := ioutil.ReadFile(filename)
	if err != nil {
		errors.Wrap(err, "Could not read input file")
	}

	// Then we open a connection to pachyderm running on localhost.
	c, err := client.NewFromAddress("localhost:30650")
	if err != nil {
		errors.Wrap(err, "Could not connect to Pachyderm")
	}

	// A repo called "godata" is created.
	if err := c.CreateRepo("godata"); err != nil {
		errors.Wrap(err, "Could not create pachyderm repo")
	}

	// We start a commit on the "master" branch of the godata repo.
	_, err = c.StartCommit("godata", "", "master")
	if err != nil {
		errors.Wrap(err, "Could not start pachyderm repo commit")
	}

	// Then after the commit is started, we can put the given file into
	// the godata repo.
	r := bytes.NewReader(csvfile)
	if _, err := c.PutFile("godata", "master", filename, r); err != nil {
		errors.Wrap(err, "Could not put file into pachyderm repo")
	}

	// Lastly, we "finish" the commit.
	if err := c.FinishCommit("godata", "master"); err != nil {
		errors.Wrap(err, "Could not finish Pachyderm commit")
	}

	return nil

}
