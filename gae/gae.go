package gae

import (
	"net/http"

	"github.com/gorilla/mux"
)

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/api/issues/{label}", HandleGetIssues).Methods("GET")

	r.HandleFunc("/tasks/issues/reset", HandleResetIssues)
	r.HandleFunc("/tasks/issues/update", HandleUpdateIssues)

	r.Handle("/", http.FileServer(http.Dir("static/dashboard")))
}
