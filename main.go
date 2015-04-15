package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tbuckley/go-issuetracker/common"
	"github.com/tbuckley/go-issuetracker/gcode"
	"github.com/tbuckley/go-issuetracker/googauth"
	"github.com/tbuckley/go-issuetracker/query"
)

var (
	fSecretsFile = flag.String("secrets", "", "Oauth secrets")
	fStorageFile = flag.String("storage", "", "Oauth storage")
	fLabel       = flag.String("label", "cr-ui-settings", "Label to filter")
)

func DisplayGroupsByIntProperty(issues []*gcode.Issue, propFunc common.IntPropertyFunc) {
	groupedIssues := common.GroupIntProperty(issues, propFunc)
	pairs := groupedIssues.PairsByValue()
	for _, pair := range pairs {
		fmt.Printf("%v: %v\n", pair.KeyString(), len(pair.Issues()))
	}
}
func DisplayGroupsByStringProperty(issues []*gcode.Issue, propFunc common.StringPropertyFunc) {
	groupedIssues := common.GroupStringProperty(issues, propFunc)
	pairs := groupedIssues.PairsByNumEntries()
	for _, pair := range pairs {
		fmt.Printf("%v: %v\n", pair.KeyString(), len(pair.Issues()))
	}
}

func main() {
	flag.Parse()

	if *fStorageFile == "" || *fSecretsFile == "" {
		fmt.Println("Usage: ./go-issuetracker --secrets=SECRETFILE --storage=STORAGEFILE")
		return
	}

	client, err := googauth.Authenticate(*fStorageFile, *fSecretsFile)
	if err != nil {
		panic(err)
	}

	log.Println("Starting requests...")

	wg := query.NewWorkGroup(20)
	q := wg.NewQuery("chromium").Client(client)
	// q = q.Label(*fLabel)
	q = q.Query("Cr:UI")

	issues := make([]*gcode.Issue, 0)
	issueChan := q.FetchAllIssues()
	for issue := range issueChan {
		if issue.Error != nil {
			fmt.Printf("Error: %v\n", issue.Error.Error())
			return
		}
		issues = append(issues, issue.Issue)
	}
	fmt.Printf("Found: %v\n", len(issues))

	priorityGroups := common.GroupIntProperty(issues, common.GetIssuePriority)
	milestoneGroups := common.GroupIntProperty(issues, common.GetIssueMilestone)
	starGroups := common.GroupIntProperty(issues, common.GetIssueStars)
	ownerGroups := common.GroupStringProperty(issues, common.GetIssueOwner)
	typeGroups := common.GroupStringProperty(issues, common.GetIssueType)
	statusGroups := common.GroupStringProperty(issues, common.GetIssueStatus)
	osGroups := common.GroupStringProperty(issues, common.GetIssueOS)
	publishedGroups := common.GroupTimeProperty(issues, common.GetIssuePublished)
	updatedGroups := common.GroupTimeProperty(issues, common.GetIssueUpdated)

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

	oldMilestoneIssues := GetOldMilestoneIssues(milestoneGroups, currentMilestone)
	fmt.Printf("Old milestones: %v\n", len(oldMilestoneIssues))

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
}

func GetOldMilestoneIssues(milestoneGroups *common.IntGroups, milestone int) []*gcode.Issue {
	issues := make([]*gcode.Issue, 0)
	milestonesSorted := milestoneGroups.Pairs()
	for _, pair := range milestonesSorted {
		intPair := pair.(*common.IntPair)
		if intPair.Key != nil && *intPair.Key < milestone {
			issues = append(issues, intPair.Entries...)
		}
	}
	return issues
}

func GetMostStarredIssue(starGroups *common.IntGroups) *gcode.Issue {
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

func GetOldestIssue(timeGroups *common.TimeGroups) *gcode.Issue {
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

func LaunchBugsForMilestone(milestoneGroups *common.IntGroups, milestone int) []*gcode.Issue {
	milestoneIssues, ok := milestoneGroups.Groups[milestone]
	if !ok {
		return nil
	}
	typeIssues := common.GroupStringProperty(milestoneIssues, common.GetIssueType)
	launchIssues, ok := typeIssues.Groups["Launch"]
	if !ok {
		return nil
	}
	return launchIssues
}
