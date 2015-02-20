package main

import (
	"strconv"
	"strings"
	"time"
)

type IntPropertyFunc func(entry *Entry) (int, bool)
type StringPropertyFunc func(entry *Entry) (string, bool)
type StringListPropertyFunc func(entry *Entry) []string
type TimePropertyFunc func(entry *Entry) (time.Time, bool)

type IntGroups struct {
	Groups map[int][]*Entry
	None   []*Entry
}

func GroupIntProperty(entries []*Entry, propFunc IntPropertyFunc) *IntGroups {
	groups := new(IntGroups)
	for _, entry := range entries {
		val, ok := propFunc(entry)
		if !ok {
			groups.None = append(groups.None, entry)
		} else {
			groups.Groups[val] = append(groups.Groups[val], entry)
		}
	}
	return groups
}

func GetIssueLabels(entry *Entry) []string {
	return entry.Labels
}

func GetIssueLabelsByPrefix(entry *Entry, prefix string) []string {
	filtered := make([]string, 0)
	labels := GetIssueLabels(entry)
	for _, label := range labels {
		if strings.HasPrefix(label, prefix) {
			filtered = append(filtered, label[len(prefix):])
		}
	}
	return labels
}

func GetIssueLabelByPrefix(entry *Entry, prefix string) (string, bool) {
	labels := GetIssueLabelsByPrefix(entry, prefix)
	if len(labels) == 1 {
		return labels[0], true
	}
	return "", false
}

func GetIssueLabelIntByPrefix(entry *Entry, prefix string) (int, bool) {
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

func GetIssuePriority(entry *Entry) (int, bool) {
	return GetIssueLabelIntByPrefix(entry, "Pri-")
}
