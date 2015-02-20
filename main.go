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
}
