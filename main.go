package main

import (
	"fmt"
)

func main() {
	q := NewQuery("chromium").Open().Limit(25)
	q = q.Label("cr-ui-settings")
	// q = q.OpenedBefore(time.Now().Add(-24 * time.Hour))

	feed, err := q.FetchPage()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	} else {
		fmt.Printf("Total issues: %v\n", feed.TotalResults)
	}

	issues, err := q.FetchAllIssues()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	} else {
		fmt.Printf("Found: %v\n", len(issues))
	}
}
