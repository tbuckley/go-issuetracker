package gae

import (
	"encoding/json"
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
	issues, err := q.FetchAllIssues()
	if err != nil {
		ctx.Errorf("Error fetching all open issues: %v", err.Error())
		return
	}
	ctx.Infof("Successfully retrieved all open issues: %v", len(issues))

	// Insert the issues
	err = AddIssues(ctx, issues)
	if err != nil {
		ctx.Errorf("Error inserting initial batch of issues: %v", err.Error())
		return
	}
	ctx.Infof("Successfully added initial batch of %v issues", len(issues))

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
	return nil, nil
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

type UpdateEntry struct {
	Updated time.Time
}

func GetUpdateKey(ctx appengine.Context) *datastore.Key {
	return datastore.NewKey(ctx, "UpdateEntry", "lastupdate", 0, nil)
}

func SetLastUpdateTime(ctx appengine.Context, updated time.Time) error {
	update := &UpdateEntry{Updated: updated}
	key := GetUpdateKey(ctx)
	_, err := datastore.Put(ctx, key, update)
	return err
}

func GetLastUpdateTime(ctx appengine.Context) (time.Time, error) {
	update := new(UpdateEntry)
	key := GetUpdateKey(ctx)
	err := datastore.Get(ctx, key, update)
	if err != nil {
		return time.Time{}, err
	}
	return update.Updated, nil
}

func DeleteLastUpdateTime(ctx appengine.Context) error {
	key := GetUpdateKey(ctx)
	return datastore.Delete(ctx, key)
}
