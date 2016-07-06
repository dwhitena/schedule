package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

const (
	REMAINING_THRESHOLD = 1
)

func main() {

	// start time
	t1 := time.Date(2016, time.July, 1, 0, 0, 0, 0, time.UTC)
	err := queryFromStartTime(t1)
	if err != nil {
		log.Fatal(err)
	}

}

// queryFromStartTime queries github for all 2 day time ranges of repo create dates
// from a start time until now
func queryFromStartTime(t1 time.Time) error {

	// Github client
	client := github.NewClient(nil)

	for t1.Unix() < time.Now().Unix() {

		// form the Github time range query
		t2 := t1.Add(time.Hour * 24 * 2)
		tString := fmt.Sprintf("\"%d-%02d-%02d .. %d-%02d-%02d\"",
			t1.Year(), t1.Month(), t1.Day(),
			t2.Year(), t2.Month(), t2.Day())
		query := fmt.Sprintf("language:Go created:" + tString)

		// execute the query using the Github client
		err := clientQuery(client, query)
		if err != nil {
			errors.Wrap(err, "Could not search Github repos")
		}

		t1 = t1.Add(time.Hour * 24 * 2)

	}

	return nil

}

// clientQuery executes github queries and searches over all pages of a result set
// parsing results
func clientQuery(gh *github.Client, query string) error {

	page := 1
	maxPage := math.MaxInt32

	opts := &github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for page <= maxPage {
		opts.Page = page
		result, response, err := gh.Search.Repositories(query, opts)
		if err != nil {
			return errors.Wrap(err, "Could not search Github result pages")
		}
		Wait(response)

		maxPage = response.LastPage
		for _, repo := range result.Repositories {

			name := *repo.FullName
			updated_at := repo.UpdatedAt.String()
			created_at := repo.CreatedAt.String()
			forks := *repo.ForksCount
			issues := *repo.OpenIssuesCount
			stars := *repo.StargazersCount
			size := *repo.Size

			fmt.Printf("%s,%s,%s,%d,%d,%d,%d\n",
				name, updated_at, created_at, forks, issues, stars, size)

		}

		time.Sleep(time.Second * 10)
		page++

	}

	return nil

}

// Wait waits to make sure we return the full github response
func Wait(response *github.Response) {
	if response != nil && response.Remaining <= REMAINING_THRESHOLD {
		gap := time.Duration(response.Reset.Local().Unix() - time.Now().Unix())
		sleep := gap * time.Second
		if sleep < 0 {
			sleep = -sleep
		}
		time.Sleep(sleep)
	}
}
