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

func DisplayGroupsByIntProperty(issues []*Entry, propFunc IntPropertyFunc) {
	groupedIssues := GroupIntProperty(issues, propFunc)
	pairs := groupedIssues.PairsByValue()
	for _, pair := range pairs {
		if pair.Key == nil {
			fmt.Printf("None: %v\n", len(pair.Entries))
		} else {
			fmt.Printf("%v: %v\n", *pair.Key, len(pair.Entries))
		}
	}
}

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

	fmt.Println("== Issues by priority ==")
	DisplayGroupsByIntProperty(issues, GetIssuePriority)
	fmt.Println("== Issues by milestone ==")
	DisplayGroupsByIntProperty(issues, GetIssueMilestone)
	fmt.Println("== Issues by stars ==")
	DisplayGroupsByIntProperty(issues, GetISsueStars)
}
