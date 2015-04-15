package reports

import (
	"appengine/taskqueue"

	"github.com/tbuckley/go-issuetracker/common"
	"github.com/tbuckley/go-issuetracker/gcode"
)

func GenerateReport(issues []*gcode.Issue) {
	// taskqueue.NewPOSTTask("/task/report", map[string][]string{
	// 	"name": {"foo"},
	// })

	starGroups := common.GroupIntProperty(issues, common.getIssueStars)
	priorityGroups := common.GroupIntProperty(issues, common.GetIssuePriority)

	report := new(Report)
	report.TotalCount = len(issues)

	report.Samples = append(report.Samples, GetGroupsIntSample(priorityGroups, 1))
}
