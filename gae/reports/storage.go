package reports

import (
	"appengine"
	"appengine/datastore"

	"github.com/tbuckley/go-issuetracker/gcode"
)

type IssuesSample struct {
	Key    string `json:"key"`
	Count  int    `json:"count"`
	Sample []int  `json:"sample"`
}

type Report struct {
	Name       string         `json:"name"`
	TotalCount int            `json:"totalCount"`
	Samples    []IssuesSample `json:"samples"`
	Layout     []Sections     `json:"sections"`
}
