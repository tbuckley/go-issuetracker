package gae

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"github.com/gorilla/mux"

	"github.com/tbuckley/go-issuetracker/gcode"
	"github.com/tbuckley/go-issuetracker/query"
)

type Response struct {
	Issues map[string]*gcode.Issue `json:"issues"`
}

func HandleGetIssues(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Get label
	vars := mux.Vars(r)
	label := vars["label"]

	// Get issues for label
	issues, err := GetAllIssuesWithLabel(ctx, label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return issues
	data, err := json.Marshal(issues)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleResetIssues(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// Delete existing issues, last update time
	err := DeleteAllIssues(ctx)
	if err != nil {
		ctx.Errorf("Error deleting all issues: %v", err.Error())
		return
	}
	ctx.Infof("Successfully deleted all existing issues")
	err = DeleteLastUpdateTime(ctx)
	if err != nil {
		ctx.Errorf("Error deleting last update time: %v", err.Error())
		return
	}
	ctx.Infof("Successfully deleted entry of last update time")

	// Get new issues
	utcNow := time.Now().UTC()
	workgroup := query.NewWorkGroup(1)
	client := urlfetch.Client(ctx)
	q := workgroup.NewQuery("chromium").Client(client)
	q = q.Label("cr-ui-settings").Open()
	issuesChan := query.BatchIssues(q.FetchAllIssues(), 25)
	for optionalIssues := range issuesChan {
		log.Printf("Handling issues!")
		if optionalIssues.Error != nil {
			ctx.Errorf("Error while fetching all open issues: %v", optionalIssues.Error.Error())
			return
		} else {
			// Insert the issues
			err = AddIssues(ctx, optionalIssues.Issues)
			if err != nil {
				ctx.Errorf("Error inserting batch of initial issues: %v", err.Error())
				return
			}
			ctx.Infof("Successfully added batch of %v initial issues", len(optionalIssues.Issues))
		}
	}
	ctx.Infof("Successfully retrieved all open issues")

	// Insert the log entry
	err = SetLastUpdateTime(ctx, utcNow)
	if err != nil {
		ctx.Errorf("Error adding an entry with the initial update time: %v", err.Error())
		return
	}
	ctx.Infof("Successfully added an entry with initial update time: %v", utcNow.Format("2006-01-02 15:04:05"))
}

func HandleUpdateIssues(w http.ResponseWriter, r *http.Request) {

}

func GetIssueKey(ctx appengine.Context, issue *gcode.Issue) *datastore.Key {
	stringID := strconv.Itoa(issue.ID)
	return datastore.NewKey(ctx, "Issue", stringID, 0, nil)
}

func GetAllIssuesWithLabel(ctx appengine.Context, label string) ([]*gcode.Issue, error) {
	q := datastore.NewQuery("Issue")
	issues := make([]*gcode.Issue, 0)
	_, err := q.Filter("Labels =", label).GetAll(ctx, &issues)
	return issues, err
}

func AddIssues(ctx appengine.Context, issues []*gcode.Issue) error {
	incompleteKeys := make([]*datastore.Key, len(issues))
	for i, issue := range issues {
		incompleteKeys[i] = GetIssueKey(ctx, issue)
	}
	_, err := datastore.PutMulti(ctx, incompleteKeys, issues)
	return err
}

func UpdateIssues(ctx appengine.Context, issues []*gcode.Issue) error {
	incompleteKeys := make([]*datastore.Key, len(issues))
	for i, issue := range issues {
		incompleteKeys[i] = GetIssueKey(ctx, issue)
	}
	_, err := datastore.PutMulti(ctx, incompleteKeys, issues)
	return err
}

func DeleteAllIssues(ctx appengine.Context) error {
	q := datastore.NewQuery("Issue")
	keys, err := q.KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return err
	}
	return datastore.DeleteMulti(ctx, keys)
}
