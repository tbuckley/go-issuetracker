package main

import (
	"sort"
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
type IntPair struct {
	Key     *int
	Entries []*Entry
}
type KeySortedIntPairList []*IntPair

func (l KeySortedIntPairList) Len() int {
	return len(l)
}
func (l KeySortedIntPairList) Less(i, j int) bool {
	switch {
	case l[i].Key == l[j].Key:
		return false
	case l[i].Key == nil:
		return true
	case l[j].Key == nil:
		return false
	default:
		return *l[i].Key < *l[j].Key
	}
}
func (l KeySortedIntPairList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (g *IntGroups) Pairs() []*IntPair {
	pairs := make([]*IntPair, 0, len(g.Groups)+1)
	for k, v := range g.Groups {
		i := k
		pairs = append(pairs, &IntPair{&i, v})
	}
	if len(g.None) > 0 {
		pairs = append(pairs, &IntPair{nil, g.None})
	}
	return pairs
}
func (g *IntGroups) PairsByValue() []*IntPair {
	pairs := g.Pairs()
	sort.Sort(KeySortedIntPairList(pairs))
	return pairs
}
func (g *IntGroups) PairsByNumEntries() []*IntPair {
	return nil
}

func GroupIntProperty(entries []*Entry, propFunc IntPropertyFunc) *IntGroups {
	groups := &IntGroups{
		Groups: make(map[int][]*Entry),
	}
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
	return filtered
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
