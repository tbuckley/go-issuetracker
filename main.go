package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tbuckley/go-issuetracker/googauth"
	"github.com/tbuckley/go-issuetracker/query"
)

var (
	secretsFile = flag.String("secrets", "", "Oauth secrets")
	storageFile = flag.String("storage", "", "Oauth storage")
)

func DisplayGroupsByIntProperty(issues []*query.Entry, propFunc IntPropertyFunc) {
	groupedIssues := GroupIntProperty(issues, propFunc)
	pairs := groupedIssues.PairsByValue()
	for _, pair := range pairs {
		fmt.Printf("%v: %v\n", pair.KeyString(), len(pair.Issues()))
	}
}
func DisplayGroupsByStringProperty(issues []*query.Entry, propFunc StringPropertyFunc) {
	groupedIssues := GroupStringProperty(issues, propFunc)
	pairs := groupedIssues.PairsByNumEntries()
	for _, pair := range pairs {
		fmt.Printf("%v: %v\n", pair.KeyString(), len(pair.Issues()))
	}
}

func main() {
	flag.Parse()

	if *storageFile == "" || *secretsFile == "" {
		fmt.Println("Usage: ./go-issuetracker --secrets=SECRETFILE --storage=STORAGEFILE")
		return
	}

	client, err := googauth.Authenticate(*storageFile, *secretsFile)
	if err != nil {
		panic(err)
	}

	log.Println("Starting requests...")

	wg := query.NewWorkGroup(20)
	q := wg.NewQuery("chromium").Client(client)
	q = q.Label("cr-ui-settings")

	issues, err := q.FetchAllIssues()
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
	} else {
		fmt.Printf("Found: %v\n", len(issues))
	}

	priorityGroups := GroupIntProperty(issues, GetIssuePriority)
	milestoneGroups := GroupIntProperty(issues, GetIssueMilestone)
	starGroups := GroupIntProperty(issues, GetISsueStars)
	ownerGroups := GroupStringProperty(issues, GetIssueOwner)
	typeGroups := GroupStringProperty(issues, GetIssueType)
	statusGroups := GroupStringProperty(issues, GetIssueStatus)
	osGroups := GroupStringProperty(issues, GetIssueOS)
	publishedGroups := GroupTimeProperty(issues, GetIssuePublished)
	updatedGroups := GroupTimeProperty(issues, GetIssueUpdated)

	fmt.Printf("Total bugs: %v\n", len(issues))

	fmt.Println("== Cleaning list ==")
	fmt.Printf("Untriaged: %v\n", len(statusGroups.Groups["Untriaged"]))
	fmt.Printf("No owner: %v\n", len(ownerGroups.None))
	fmt.Printf("No milestone: %v\n", len(milestoneGroups.None))
	fmt.Printf("No priority: %v\n", len(priorityGroups.None))
	fmt.Printf("No type: %v\n", len(typeGroups.None))
	fmt.Printf("No status: %v\n", len(statusGroups.None))
	fmt.Printf("No OS: %v\n", len(osGroups.None))

	currentMilestone := 42

	milestonesSorted := milestoneGroups.Pairs()
	oldMilestones := 0
	for _, pair := range milestonesSorted {
		intPair := pair.(*IntPair)
		if intPair.Key != nil && *intPair.Key < currentMilestone {
			oldMilestones += len(intPair.Entries)
		}
	}
	fmt.Printf("Old milestones: %v\n", oldMilestones)

	fmt.Println("== Priority list ==")
	fmt.Printf("P1: %v\n", len(priorityGroups.Groups[1]))
	milestoneLaunchBugs := LaunchBugsForMilestone(milestoneGroups, currentMilestone)
	fmt.Printf("M%v Launch bugs: %v\n", currentMilestone, len(milestoneLaunchBugs))
	nextMilestoneLaunchBugs := LaunchBugsForMilestone(milestoneGroups, currentMilestone+1)
	fmt.Printf("M%v Launch bugs: %v\n", currentMilestone+1, len(nextMilestoneLaunchBugs))

	fmt.Println("== Superlatives list ==")
	mostStarred := GetMostStarredIssue(starGroups)
	fmt.Printf("Top stars: crbug.com/%v (%v)\n", mostStarred.ID, mostStarred.Stars)
	lastPublished := GetOldestIssue(publishedGroups)
	fmt.Printf("Oldest published: crbug.com/%v (%v)\n", lastPublished.ID, lastPublished.Published)
	lastUpdated := GetOldestIssue(updatedGroups)
	fmt.Printf("Oldest updated: crbug.com/%v (%v)\n", lastUpdated.ID, lastUpdated.Published)

	// fmt.Println("== Issues by priority ==")
	// DisplayGroupsByIntProperty(issues, GetIssuePriority)
	// fmt.Println("== Issues by milestone ==")
	// DisplayGroupsByIntProperty(issues, GetIssueMilestone)
	// fmt.Println("== Issues by stars ==")
	// DisplayGroupsByIntProperty(issues, GetISsueStars)
	// fmt.Println("== Issues by owner ==")
	// DisplayGroupsByStringProperty(issues, GetIssueOwner)
	// fmt.Println("== Issues by type ==")
	// DisplayGroupsByStringProperty(issues, GetIssueType)
	// fmt.Println("== Issues by status ==")
	// DisplayGroupsByStringProperty(issues, GetIssueStatus)
}

func GetMostStarredIssue(starGroups *IntGroups) *query.Entry {
	starsSorted := starGroups.PairsByValue()
	if len(starsSorted) == 0 {
		return nil
	}
	lastBucketIndex := len(starsSorted) - 1
	lastBucket := starsSorted[lastBucketIndex].Issues()
	if len(lastBucket) == 0 {
		return nil
	}
	return lastBucket[0]
}

func GetOldestIssue(timeGroups *TimeGroups) *query.Entry {
	timesSorted := timeGroups.PairsByValue()
	if len(timesSorted) == 0 {
		return nil
	}
	oldestBucketIndex := 0
	oldestBucket := timesSorted[oldestBucketIndex].Issues()
	if len(oldestBucket) == 0 {
		return nil
	}
	return oldestBucket[0]
}

func LaunchBugsForMilestone(milestoneGroups *IntGroups, milestone int) []*query.Entry {
	milestoneIssues, ok := milestoneGroups.Groups[milestone]
	if !ok {
		return nil
	}
	typeIssues := GroupStringProperty(milestoneIssues, GetIssueType)
	launchIssues, ok := typeIssues.Groups["Launch"]
	if !ok {
		return nil
	}
	return launchIssues
}
