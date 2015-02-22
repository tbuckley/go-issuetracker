package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/tbuckley/go-issuetracker/query"
)

func GetIssueLabels(entry *query.Entry) []string {
	return entry.Labels
}

func GetIssueLabelsByPrefix(entry *query.Entry, prefix string) []string {
	filtered := make([]string, 0)
	labels := GetIssueLabels(entry)
	for _, label := range labels {
		if strings.HasPrefix(label, prefix) {
			filtered = append(filtered, label[len(prefix):])
		}
	}
	return filtered
}

func GetIssueLabelByPrefix(entry *query.Entry, prefix string) (string, bool) {
	labels := GetIssueLabelsByPrefix(entry, prefix)
	if len(labels) == 1 {
		return labels[0], true
	}
	return "", false
}

func GetIssueLabelIntByPrefix(entry *query.Entry, prefix string) (int, bool) {
	priorityString, ok := GetIssueLabelByPrefix(entry, prefix)
	if !ok {
		return 0, false
	}
	priority, err := strconv.Atoi(priorityString)
	if err != nil {
		return 0, false
	}
	return priority, true
}

func GetIssueCrLabels(entry *query.Entry) []string {
	return GetIssueLabelsByPrefix(entry, "Cr-")
}

func GetIssuePriority(entry *query.Entry) (int, bool) {
	return GetIssueLabelIntByPrefix(entry, "Pri-")
}

func GetIssueMilestone(entry *query.Entry) (int, bool) {
	return GetIssueLabelIntByPrefix(entry, "M-")
}

func GetISsueStars(entry *query.Entry) (int, bool) {
	return entry.Stars, true
}

func GetIssueOwner(entry *query.Entry) (string, bool) {
	if entry.Owner != nil {
		return *entry.Owner, true
	}
	return "", false
}

func GetIssueStatus(entry *query.Entry) (string, bool) {
	return entry.Status, len(entry.Status) > 0
}

func GetIssueType(entry *query.Entry) (string, bool) {
	return GetIssueLabelByPrefix(entry, "Type-")
}

func GetIssueOS(entry *query.Entry) (string, bool) {
	return GetIssueLabelByPrefix(entry, "OS-")
}

func GetIssuePublished(entry *query.Entry) (time.Time, bool) {
	// 2015-02-18T00:36:15.000Z
	// Mon Jan 2 15:04:05 -0700 MST 2006
	parsed, err := time.Parse("2006-01-02T15:04:05.000Z", entry.Published)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func GetIssueUpdated(entry *query.Entry) (time.Time, bool) {
	// 2015-02-18T00:36:15.000Z
	// Mon Jan 2 15:04:05 -0700 MST 2006
	parsed, err := time.Parse("2006-01-02T15:04:05.000Z", entry.Updated)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}
