package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	secretsFile = flag.String("secrets", "", "Oauth secrets")
	storageFile = flag.String("storage", "", "Oauth storage")
)

func main() {
	flag.Parse()

	client, err := Authenticate(*storageFile, *secretsFile)
	if err != nil {
		panic(err)
	}

	log.Println("Starting requests...")

	q := NewQuery("chromium").Client(client)
	q = q.Label("cr-ui-settings")

	issues, err := q.FetchAllIssues()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	} else {
		fmt.Printf("Found: %v\n", len(issues))
	}

	issuesByPriority := GroupIntProperty(issues, GetIssuePriority)
	pairs := issuesByPriority.PairsByValue()
	for _, pair := range pairs {
		if pair.Key == nil {
			fmt.Printf("None: %v\n", len(pair.Entries))
		} else {
			fmt.Printf("%v: %v\n", *pair.Key, len(pair.Entries))
		}
	}
}
